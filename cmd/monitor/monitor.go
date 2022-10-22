package monitor

import (
	"net/http"
	"os"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/synycboom/algorand-notification/fetcher"
)

var (
	configFile string

	Command = &cobra.Command{
		Use: "monitor",
		Short: "run monitor daemon",
		Long: "run monitor daemon that watches new blocks and publishes events.",
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
  Command.MarkFlagRequired("config")
	cobra.OnInitialize(initConfig)
}

// initConfig initialize configuration
func initConfig() {
  viper.SetConfigType("yaml")
  viper.SetConfigFile(configFile)

  if err := viper.ReadInConfig(); err != nil {
    log.Error().Err(err).Msg("initConfig: invalid configfile")
    os.Exit(1)
  } else {
    log.Info().Msgf("initConfig: using config file %s", viper.ConfigFileUsed())
  }
}

func run() error {
	indexerHost := viper.GetString("INDEXER_HOST")
	indexerAPIToken := viper.GetString("INDEXER_API_TOKEN")
	fetcherRPS := viper.GetInt("FETCHER_RPS")
	startRound := viper.GetUint64("START_ROUND")
	port := viper.GetString("PORT")

	f, err := fetcher.New(fetcher.Config{
		Host:       indexerHost,
		APIToken:   indexerAPIToken,
		RPS:        fetcherRPS,
		StartRound: startRound,
		Processor: func(b *models.Block) {
			log.Info().Msgf("%v", b.Round)
		},
	})
	if err != nil {
		return err
	}

	f.Start()
	defer f.Stop()

	log.Info().Msg("monitor is running on port " + port)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		return err
	}

	return nil
}
