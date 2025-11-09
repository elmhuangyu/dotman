package cmd

import (
	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dotfiles to the system",
	Long: `Install dotfiles from the configured dotfiles directory to the system.
This command copies and links configuration files to their appropriate locations.`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.GetLogger()
		log.Info().Msg("Install command executed (not yet implemented)")
	},
}