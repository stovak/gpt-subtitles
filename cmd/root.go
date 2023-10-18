/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stovak/gpt-subtitles/pkg/models"
	"log"
	"os"

	"github.com/ayush6624/go-chatgpt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var gptClient *chatgpt.Client = GetGPTClient()

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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gpt-subtitles",
	Short: "Translate a subtitle file using GPT-4",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("must provide a subtitle file")
		}
		subtitleFile := args[0]
		tr := models.NewTranslationRequestFromFile(subtitleFile, "en", "es")
		buf := new(bytes.Buffer)
		err := tr.ToPrompt(buf)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(buf.String())
		res, err := gptClient.Send(context.Background(), &chatgpt.ChatCompletionRequest{
			Model: chatgpt.GPT4,
			Messages: []chatgpt.ChatMessage{
				{
					Role:    chatgpt.ChatGPTModelRoleSystem,
					Content: buf.String(),
				},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res.Object.Choices[0].Message)
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

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
