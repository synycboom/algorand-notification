package server

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/synycboom/algorand-notification/client"
	"github.com/synycboom/algorand-notification/event"
	"github.com/synycboom/algorand-notification/handler"
	"github.com/synycboom/algorand-notification/hub"
	"github.com/synycboom/algorand-notification/subscriber"
)

var (
	configFile string
	Command = &cobra.Command{
		Use:   "server",
		Short: "run server",
		Long:  "run server that subscribes events and accepts websocket connections.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(); err != nil {
				log.Error().Err(err).Msg("server: unexpected error")
				os.Exit(1)
			}
		},
	}
)

func init() {
	flags := Command.Flags()
	flags.StringVarP(&configFile, "config", "c", "", "file path to configuration file (server.yml)")

  if err := Command.MarkFlagRequired("config"); err != nil {
    os.Exit(1)
  }
}

func run() error {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	log.Info().Msgf("server: using config file %s", viper.ConfigFileUsed())

	port := viper.GetString("PORT")
	metricsPort := viper.GetString("METRICS_PORT")
	redisHost := viper.GetString("REDIS_HOST")
	redisPassword := viper.GetString("REDIS_PASSWORD")
	channel := viper.GetString("NEW_BLOCK_CHANNEL")
	logLevel, err := zerolog.ParseLevel(viper.GetString("LOG_LEVEL"))
	if err == nil {
		zerolog.SetGlobalLevel(logLevel)
	}

	h, err := hub.New(30000)
	if err != nil {
		return err
	}
	defer h.Close()

	pingInterval, err := time.ParseDuration("3m")
	if err != nil {
		return err
	}

	pongWait, err := time.ParseDuration("3m15s")
	if err != nil {
		return err
	}

	f, err := client.NewFactory(client.Config{
		WriteWaitTimeout:   time.Duration(5) * time.Second,
		PongWaitTimeout:    pongWait,
		PingInterval:       pingInterval,
		MaxReadMessageSize: 1024,
		SendBufferSize:     100,
	})
	if err != nil {
		return nil
	}

	hnd := handler.New(handler.Config{
		Hub: h,
		Upgrader: &websocket.Upgrader{
			EnableCompression: false,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		ClientFactory: f,
	})

	echoMainServer := echo.New()
	echoMainServer.HideBanner = true
	echoMainServer.Use(middleware.Logger())
	echoMainServer.GET("/", hnd.Upgrade)

	echoPrometheus := echo.New()
	echoPrometheus.HideBanner = true
	prom := prometheus.NewPrometheus("echo", nil)

	echoMainServer.Use(prom.HandlerFunc)
	prom.SetMetricsPath(echoPrometheus)

	s, err := subscriber.NewRedis(&subscriber.RedisConfig{
		RedisHost:     redisHost,
		RedisPassword: redisPassword,
		Channel:       channel,
		Processor: func(data []byte) {
			events, err := event.Parse(data)
			if err != nil {
				log.Error().Err(err).Msg("server: failed to parse an event")
			}

			for _, event := range events {
				h.SendEvent(event)
			}
		},
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Error().Err(err).Msg("server: unexpected error")
			os.Exit(1)
		}
	}()

	go h.Run()
	go func() {
		if err := echoPrometheus.Start(":" + metricsPort); err != nil {
			log.Error().Err(err).Msg("server-metrics: unexpected error")
			os.Exit(1)
		}
	}()

	if err := echoMainServer.Start(":" + port); err != nil {
		return err
	}

	return nil
}
