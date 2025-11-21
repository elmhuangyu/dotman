package cmd

import (
	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dotfiles to the system",
	Long: `Install dotfiles from the configured dotfiles directory to the system.
This command copies and links configuration files to their appropriate locations.`,
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

		// TODO: Implement installation logic using cfg.Modules
		for _, module := range cfg.Modules {
			log.Info().Str("dir", module.Dir).Str("target_dir", module.TargetDir).Msg("Found module")
		}

		log.Info().Msg("Install command completed")
		return nil
	},
}
