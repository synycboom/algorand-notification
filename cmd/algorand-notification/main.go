package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
  log.Info().Msgf("indexer_host: %s", viper.GetString("INDEXER_HOST"))
  log.Info().Msgf("port: %s", viper.GetString("PORT"))
  log.Info().Msgf("start_round: %s", viper.GetUint64("START_ROUND"))
  log.Info().Msgf("fetcher_rps: %s", viper.GetInt("FETCHER_RPS"))

  return nil
}
