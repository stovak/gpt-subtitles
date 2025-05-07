/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/stovak/gpt-subtitles/cmd/drop"
	"github.com/stovak/gpt-subtitles/cmd/subs"
)

var (
	Logger      *zap.SugaredLogger = initLogger().Sugar()
	enableDebug                    = false
)

var cfgFile string

func initLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	if enableDebug {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()
	return logger
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "subtitles",
	Short: "Translate a subtitle file using GPT-4 or Google Translate",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Display help messages from all commands
		return cmd.Help()
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
	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gpt-subtitles.yaml)")
	rootCmd.PersistentFlags().StringP("sourceLanguage", "s", "en", "SourceLanguage... E.g. en for English")
	rootCmd.PersistentFlags().StringP("targetLanguage", "t", "es", "DestinationLanguage... E.g. es for Spanish")
	rootCmd.PersistentFlags().StringP("engine", "e", "gpt", "Translation Engine: google or gpt")
	rootCmd.PersistentFlags().BoolVar(&enableDebug, "debug", os.Getenv("DEBUG") == "true", "Enable debug mode")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	rootCmd.AddCommand(subs.TranslateAllCmd)
	rootCmd.AddCommand(subs.TranslateAllCmd)
	rootCmd.AddCommand(drop.ListCmd)

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".gpt-subtitles" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".subtitles")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		Logger.Infof("Using config file: %s", viper.ConfigFileUsed())
	}
}
