package cmd

import (
	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall dotfiles from the system",
	Long: `Remove dotfiles from the system and restore original files if available.
This command cleans up configuration files installed by the install command.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()

		dotfilesDir := getDotfilesDir()
		log.Info().Str("dotfiles_dir", dotfilesDir).Msg("Loading configuration")

		cfg, err := config.LoadDir(dotfilesDir)
		if err != nil {
			return err
		}

		log.Info().Int("modules", len(cfg.Modules)).Msg("Configuration loaded successfully")

		// TODO: Implement uninstallation logic using cfg.Modules
		for _, module := range cfg.Modules {
			log.Info().Str("dir", module.Dir).Str("target_dir", module.TargetDir).Msg("Found module to uninstall")
		}

		log.Info().Msg("Uninstall command completed")
		return nil
	},
}
