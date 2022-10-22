package server

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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

	Command.MarkFlagRequired("config")
}

// initConfig initialize configuration
func initConfig() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("server: invalid configfile")
		os.Exit(1)
	} else {
		log.Info().Msgf("server: using config file %s", viper.ConfigFileUsed())
	}
}

func run() error {
  initConfig()

	port := viper.GetString("PORT")
	redisHost := viper.GetString("REDIS_HOST")
	redisPassword := viper.GetString("REDIS_PASSWORD")
	channel := viper.GetString("NEW_BLOCK_CHANNEL")
	s, err := subscriber.NewRedis(&subscriber.RedisConfig{
		RedisHost:     redisHost,
		RedisPassword: redisPassword,
		Channel:       channel,
    Processor: func(message []byte) {
      log.Info().Msgf("%#v", string(message))
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

  log.Info().Msg("server: running on port " + port)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		return err
	}

	return nil
}
