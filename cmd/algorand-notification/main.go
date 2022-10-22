package main

import (
	"net/http"
	"os"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/synycboom/algorand-notification/fetcher"
)

var (
	configFile string
)

var rootCmd = &cobra.Command{
	Use:   "algorand-notification",
	Short: "Algorand Notification Service",
	Long:  "Algorand notification is a notification service for Algorand Blockchain which provides websocket interface for consumers",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "run algorand-notification daemon",
	Long:  "run algorand-notification daemon. Serve websocket.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDaemon(); err != nil {
			log.Error().Err(err).Msg("daemon: unexpected error")
			os.Exit(1)
		}
	},
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	cobra.OnInitialize(initConfig)
	flags := daemonCmd.Flags()
	flags.StringVarP(&configFile, "configfile", "c", "", "file path to configuration file (config.yml)")
	daemonCmd.MarkFlagRequired("configfile")
	rootCmd.AddCommand(daemonCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("main: unexpected error")
		os.Exit(1)
	}
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigType("yaml")
		viper.SetConfigFile(configFile)

		if err := viper.ReadInConfig(); err != nil {
			log.Error().Err(err).Msg("initConfig: invalid configfile")
		} else {
			log.Info().Msgf("initConfig: using config file %s", viper.ConfigFileUsed())
		}
	}
}

func runDaemon() error {
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

	log.Info().Msg("daemon is running on port " + port)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		return err
	}

	return nil
}
