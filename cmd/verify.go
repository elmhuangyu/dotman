package cmd

import (
	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify dotfile installation status",
	Long: `Check the status of installed dotfiles and verify they are correctly linked.
This command reports on the current state of your dotfile configuration.`,
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

		// TODO: Implement verification logic using cfg.Modules
		for _, module := range cfg.Modules {
			log.Info().Str("dir", module.Dir).Str("target_dir", module.TargetDir).Msg("Found module to verify")
		}

		log.Info().Msg("Verify command completed")
		return nil
	},
}
