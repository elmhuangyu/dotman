package cmd

import (
	"fmt"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	dryRunFlag     bool
	forceFlag      bool
	interactiveFlag bool
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dotfiles to the system",
	Long: `Install dotfiles from the configured dotfiles directory to the system.
This command copies and links configuration files to their appropriate locations.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate mutually exclusive flags
		flagCount := 0
		if dryRunFlag {
			flagCount++
		}
		if forceFlag {
			flagCount++
		}
		if interactiveFlag {
			flagCount++
		}

		if flagCount > 1 {
			return fmt.Errorf("only one of --dry-run, --force, or --interactive can be used at a time")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()

		// Log which mode we're running in
		if dryRunFlag {
			log.Info().Msg("Running in dry-run mode - no changes will be made")
		} else if forceFlag {
			log.Info().Msg("Running in force mode - existing files will be overwritten")
		} else if interactiveFlag {
			log.Info().Msg("Running in interactive mode - will prompt for confirmation")
		}

		dotfilesDir := getDotfilesDir()
		log.Info().Str("dotfiles_dir", dotfilesDir).Msg("Loading configuration")

		cfg, err := config.LoadDir(dotfilesDir)
		if err != nil {
			return err
		}

		log.Info().Int("modules", len(cfg.Modules)).Msg("Configuration loaded successfully")

		// TODO: Implement installation logic using cfg.Modules with the new flags
		for _, module := range cfg.Modules {
			log.Info().Str("dir", module.Dir).Str("target_dir", module.TargetDir).Msg("Found module")
		}

		log.Info().Msg("Install command completed")
		return nil
	},
}

func init() {
	installCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show what would be installed without making changes")
	installCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force installation by overwriting existing files")
	installCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "Prompt before installing each file")
}
