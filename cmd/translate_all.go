/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stovak/gpt-subtitles/pkg/actions"
	"github.com/stovak/gpt-subtitles/pkg/models"
	"maps"
)

// translate:allCmd represents the translate:all command
var translateAllCmd = &cobra.Command{
	Use:   "translate:all",
	Short: "Translate a base subtitle file using GPT-4 or Google Translate into all the languages available in the config file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var tr models.TranslationRequest
		Logger.Info("Root Command Exec:")
		source, err := cmd.Flags().GetString("sourceLanguage")
		if err != nil {
			return err
		}
		engine, err := cmd.Flags().GetString("engine")
		if err != nil {
			return err
		}

		var langCopy map[string]string
		// 1. clone models.Languages
		langCopy = maps.Clone(models.Languages)
		// 2. remove the source language from the map
		delete(langCopy, source)
		// 3. for each language in the list, create a new translation request and send it to the translation engine
		for k, _ := range langCopy {
			switch engine {
			case "google":
				Logger.Info("Using Google Translate")
				tr, err = models.NewGoogleTranslationRequestFromFile(args[0], source, k, Logger)
				break
			case "gpt":
				Logger.Info("Using GPT Translate")
				tr, err = models.NewGPTTranslationRequestFromFile(args[0], source, k, Logger)
				break
			default:
				Logger.Fatalf("Unknown engine %s", engine)
			}
			err := actions.TranslateOne(tr)
			if err != nil {
				Logger.Warnf("Error translating %s => %s: %s", source, k, err)
			}
		}

		return nil
	},
}
