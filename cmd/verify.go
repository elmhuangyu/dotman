package cmd

import (
	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify dotfile installation status",
	Long: `Check the status of installed dotfiles and verify they are correctly linked.
This command reports on the current state of your dotfile configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.GetLogger()
		log.Info().Msg("Verify command executed (not yet implemented)")
	},
}