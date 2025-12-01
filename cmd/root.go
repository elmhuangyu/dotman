package cmd

import (
	"fmt"
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
It provides commands to install and uninstall dotfile configurations.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set debug mode if flag is provided
		if debugFlag {
			logger.SetDebugMode()
		}

		// Log startup info
		log := logger.GetLogger()
		_, err := getDotfilesDir()
		if err != nil {
			log.Error().Err(err).Msg("Failed to determine dotfiles directory")
			return err
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log := logger.GetLogger()
		log.Error().Msg(err.Error())
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
}

// getDotfilesDir returns the dotfiles directory based on flag or default
func getDotfilesDir() (string, error) {
	if dirFlag != "" {
		return dirFlag, nil
	}
	dir := getDefaultDotfilesDir()
	if dir == "" {
		return "", fmt.Errorf("no dotfiles directory found: neither ~/dotfiles nor ~/.config/dotfiles exist")
	}
	return dir, nil
}

// getDefaultDotfilesDir returns the default dotfiles directory
func getDefaultDotfilesDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/dotfiles"
	}

	// Check if ~/dotfiles exists
	dotfilesDir := filepath.Join(home, "dotfiles")
	if info, err := os.Stat(dotfilesDir); err == nil && info.IsDir() {
		return dotfilesDir
	}

	// Check if ~/.config/dotfiles exists
	configDir := filepath.Join(home, ".config", "dotfiles")
	if info, err := os.Stat(configDir); err == nil && info.IsDir() {
		return configDir
	}

	return ""
}
