package cmd

import (
	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall dotfiles from the system",
	Long: `Remove dotfiles from the system and restore original files if available.
This command cleans up configuration files installed by the install command.`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.GetLogger()
		log.Info().Msg("Uninstall command executed (not yet implemented)")
	},
}