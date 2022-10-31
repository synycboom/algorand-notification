package monitor

import (
	"context"
	"encoding/json"
	"os"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/synycboom/algorand-notification/fetcher"
	"github.com/synycboom/algorand-notification/publisher"
)

var (
	configFile string

	Command = &cobra.Command{
		Use:   "monitor",
		Short: "run monitor daemon",
		Long:  "run monitor daemon that watches new blocks and publishes events.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(); err != nil {
				log.Error().Err(err).Msg("daemon: unexpected error")
				os.Exit(1)
			}
		},
	}
)

func init() {
	flags := Command.Flags()
	flags.StringVarP(&configFile, "config", "c", "", "file path to configuration file (monitor.yml)")

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

	indexerHost := viper.GetString("INDEXER_HOST")
	indexerAPIToken := viper.GetString("INDEXER_API_TOKEN")
	fetcherRPS := viper.GetInt("FETCHER_RPS")
	metricsPort := viper.GetString("METRICS_PORT")
	redisHost := viper.GetString("REDIS_HOST")
	redisPassword := viper.GetString("REDIS_PASSWORD")
	publishTimeout := viper.GetDuration("PUBLISHER_TIMEOUT")
	channel := viper.GetString("NEW_BLOCK_CHANNEL")
	var startRound *uint64 = nil
	if viper.GetString("START_ROUND") != "latest" {
		value := viper.GetUint64("START_ROUND")
		startRound = &value
	}

	logLevel, err := zerolog.ParseLevel(viper.GetString("LOG_LEVEL"))
	if err == nil {
		zerolog.SetGlobalLevel(logLevel)
	}

	p, err := publisher.NewRedis(publisher.RedisConfig{
		RedisHost:     redisHost,
		RedisPassword: redisPassword,
		Channel:       channel,
	})
	if err != nil {
		return err
	}

	f, err := fetcher.New(fetcher.Config{
		Host:       indexerHost,
		APIToken:   indexerAPIToken,
		RPS:        fetcherRPS,
		StartRound: startRound,
		Processor: func(b *models.Block) {
			ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
			defer cancel()

			message, err := json.Marshal(b)
			if err != nil {
				log.Error().Err(err).Msg("monitor: failed to marshal json")
			}

			if err := p.Publish(ctx, message); err != nil {
				log.Error().Err(err).Msg("monitor: failed to publish a block")
			}
		},
	})
	if err != nil {
		return err
	}

	f.Start()
	defer f.Stop()

	echoPrometheus := echo.New()
	echoPrometheus.HideBanner = true
	prom := prometheus.NewPrometheus("echo", nil)
	prom.SetMetricsPath(echoPrometheus)
	if err := echoPrometheus.Start(":" + metricsPort); err != nil {
		return err
	}

	return nil
}
