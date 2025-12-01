package cmd

import (
	"fmt"

	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/elmhuangyu/dotman/pkg/module"
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
		dotfilesDir := getDotfilesDir()
		return uninstall(dotfilesDir)
	},
}

// uninstall performs the dotfiles uninstallation
func uninstall(dotfilesDir string) error {
	log := logger.GetLogger()

	log.Info().Str("dotfiles_dir", dotfilesDir).Msg("Starting uninstallation")

	// Perform uninstallation using the uninstall module
	result, err := module.Uninstall(dotfilesDir)
	if err != nil {
		return fmt.Errorf("uninstall failed: %w", err)
	}

	// Log the results
	log.Info().Str("summary", result.Summary).Msg("Uninstall completed")

	// Log any errors that occurred during the process
	if len(result.Errors) > 0 {
		log.Warn().Int("error_count", len(result.Errors)).Msg("Errors occurred during uninstall")
		for _, errorMsg := range result.Errors {
			log.Warn().Str("error", errorMsg).Msg("Uninstall error")
		}
	}

	// Log skipped links with reasons
	if len(result.SkippedLinks) > 0 {
		log.Info().Int("skipped_count", len(result.SkippedLinks)).Msg("Some links were skipped")
		for _, skipped := range result.SkippedLinks {
			log.Info().
				Str("target", skipped.Operation.Target).
				Str("reason", skipped.Reason).
				Msg("Skipped symlink removal")
		}
	}

	// Log skipped generated files with reasons
	if len(result.SkippedGenerated) > 0 {
		log.Info().Int("skipped_count", len(result.SkippedGenerated)).Msg("Some generated files were skipped")
		for _, skipped := range result.SkippedGenerated {
			log.Info().
				Str("target", skipped.Operation.Target).
				Str("reason", skipped.Reason).
				Msg("Skipped generated file removal")
		}
	}

	// Log backed up generated files
	if len(result.BackedUpGenerated) > 0 {
		log.Warn().Int("backed_up_count", len(result.BackedUpGenerated)).Msg("Some generated files were backed up due to modifications")
		for _, backedUp := range result.BackedUpGenerated {
			log.Warn().
				Str("target", backedUp.Operation.Target).
				Str("reason", backedUp.Reason).
				Msg("Backed up modified generated file")
		}
	}

	// Log failed removals with reasons
	if len(result.FailedRemovals) > 0 {
		log.Error().Int("failed_count", len(result.FailedRemovals)).Msg("Some files failed to remove")
		for _, failed := range result.FailedRemovals {
			log.Error().
				Str("target", failed.Operation.Target).
				Str("reason", failed.Reason).
				Msg("Failed file removal")
		}
	}

	if !result.IsSuccess {
		return fmt.Errorf("uninstall completed with errors: %s", result.Summary)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
