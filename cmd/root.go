/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stovak/gpt-subtitles/pkg/models"
	"go.uber.org/zap"
	"os"
	"strings"
)

var (
	Logger *zap.SugaredLogger = initLogger().Sugar()
)

var cfgFile string

func initLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	return logger
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "subtitles",
	Short: "Translate a subtitle file using GPT-4",
	RunE: func(cmd *cobra.Command, args []string) error {
		var tr models.TranslationRequest
		Logger.Info("Root Command Exec:")
		source, err := cmd.Flags().GetString("sourceLanguage")
		if err != nil {
			return err
		}
		dest, err := cmd.Flags().GetString("targetLanguage")
		if err != nil {
			return err
		}
		engine, err := cmd.Flags().GetString("engine")
		if err != nil {
			return err
		}
		switch engine {
		case "google":
			Logger.Info("Using Google Translate")
			tr, err = models.NewGoogleTranslationRequestFromFile(args[0], source, dest, Logger)
			break
		case "gpt":
			Logger.Info("Using GPT Translate")
			tr, err = models.NewGPTTranslationRequestFromFile(args[0], source, dest, Logger)
			break
		default:
			Logger.Fatalf("Unknown engine %s", engine)
		}
		if err != nil {
			return err
		}
		err = tr.Parse()
		if err != nil {
			return err
		}
		err = tr.Translate()
		if err != nil {
			return err
		}
		translated, err := tr.GetTranslated()
		if err != nil {
			return err
		}
		buf := new(strings.Builder)
		err = translated.WriteToTTML(buf)
		Logger.Debugf("Translated: %s", buf.String())
		return err
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
	rootCmd.PersistentFlags().StringP("engine", "e", "google", "Translation Engine: google or gpt")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
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
