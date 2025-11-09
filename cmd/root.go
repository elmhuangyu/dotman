package cmd

import (
	"os"
	"path/filepath"

	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	debugFlag bool
	dirFlag   string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dotman",
	Short: "A dotfile management and installation tool",
	Long: `dotman is a CLI tool for managing and installing dotfiles.
It provides commands to install, uninstall, and verify dotfile configurations.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger.Init()

		// Set debug mode if flag is provided
		if debugFlag {
			logger.SetDebugMode()
		}

		// Log startup info
		log := logger.GetLogger()
		log.Debug().Str("dotfiles_dir", getDotfilesDir())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&dirFlag, "dir", "", "Custom dotfiles directory (default: $HOME/.config/dotfiles)")

	// Add subcommands
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(verifyCmd)
}

// getDotfilesDir returns the dotfiles directory based on flag or default
func getDotfilesDir() string {
	if dirFlag != "" {
		return dirFlag
	}
	return getDefaultDotfilesDir()
}

// getDefaultDotfilesDir returns the default dotfiles directory
func getDefaultDotfilesDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/dotfiles"
	}
	return filepath.Join(home, ".config", "dotfiles")
}

