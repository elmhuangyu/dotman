package cmd

import (
	"fmt"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/elmhuangyu/dotman/pkg/module"
	"github.com/spf13/cobra"
)

var (
	dryRunFlag bool
	forceFlag  bool
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
		if dryRunFlag && forceFlag {
			return fmt.Errorf("only one of --dry-run or --force can be used at a time")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		dotfilesDir := getDotfilesDir()
		return install(dotfilesDir, dryRunFlag, forceFlag)
	},
}

// install performs the dotfiles installation
func install(dotfilesDir string, dryRun, force bool) error {
	log := logger.GetLogger()

	// Log which mode we're running in
	if dryRun {
		log.Info().Msg("Running in dry-run mode - no changes will be made")
	} else if force {
		log.Info().Msg("Running in force mode - existing files will be overwritten")
	}

	log.Info().Str("dotfiles_dir", dotfilesDir).Msg("Loading configuration")

	cfg, err := config.LoadDir(dotfilesDir)
	if err != nil {
		return err
	}

	log.Info().Int("modules", len(cfg.Modules)).Msg("Configuration loaded successfully")

	if !force {
		// Perform dry-run validation
		result, err := module.Validate(cfg.Modules)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		// Log the results
		module.LogValidateResult(result)

		// Return error if validation failed
		if !result.IsValid {
			return fmt.Errorf("validation failed with %d errors and %d conflicts", len(result.Errors), len(result.ConflictOperations))
		}

		if dryRun {
			log.Info().Msg("Dry-run completed successfully - no changes were made")
			return nil
		}
	}
	// TODO: Implement actual installation logic
	log.Info().Msg("Installation logic not yet implemented")
	return fmt.Errorf("actual installation is not yet implemented - use --dry-run to validate configuration")
}

func init() {
	installCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show what would be installed without making changes")
	installCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force installation by overwriting existing files")
}
