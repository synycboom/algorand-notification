package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/synycboom/algorand-notification/cmd/monitor"
)

var rootCmd = &cobra.Command{
	Use:   "algorand-notification",
	Short: "Algorand Notification Service",
	Long:  "Algorand notification is a notification service for Algorand Blockchain which provides websocket interface for consumers",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	rootCmd.AddCommand(monitor.Command)
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("main: unexpected error")
		os.Exit(1)
	}
}

