/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/stovak/gpt-subtitles/pkg/models"
	"go.uber.org/zap"
	"log"
	"os"

	"github.com/ayush6624/go-chatgpt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	GptClient *chatgpt.Client    = GetGPTClient()
	Logger    *zap.SugaredLogger = initLogger().Sugar()
)

var cfgFile string

func GetGPTClient() *chatgpt.Client {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}
	client, err := chatgpt.NewClient(key)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

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
		Logger.Info("Root Command Exec:")
		source := cmd.Flag("sourceLanguage").Value.String()
		dest := cmd.Flag("destinationLanguage").Value.String()
		if len(args) < 1 {
			return fmt.Errorf("must provide a subtitle file")
		}
		subtitleFile := args[0]
		tr, err := models.NewTranslationRequestFromFile(subtitleFile, source, dest, Logger)
		if err != nil {
			return err
		}
		res, err := GptClient.Send(context.Background(), &chatgpt.ChatCompletionRequest{
			Model: chatgpt.GPT4,
			Messages: []chatgpt.ChatMessage{
				{
					Role:    chatgpt.ChatGPTModelRoleSystem,
					Content: tr.String(),
				},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		tr.WriteResultsToFile(res)
		return nil
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
	rootCmd.PersistentFlags().StringP("destinationLanguage", "d", "es", "DestinationLanguage... E.g. es for Spanish")

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
		viper.SetConfigName(".gpt-subtitles")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		Logger.Infof("Using config file: %s", viper.ConfigFileUsed())
	}
}
