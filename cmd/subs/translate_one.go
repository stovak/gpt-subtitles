/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package subs

import (
	"github.com/stovak/gpt-subtitles/pkg/actions"
	"github.com/stovak/gpt-subtitles/pkg/models"

	"github.com/spf13/cobra"
)

// translate:oneCmd represents the translate:one command
var translateOneCmd = &cobra.Command{
	Use:   "translate:one",
	Short: "Translate a given subtitle file into a single language",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var tr models.TranslationRequest
		cmd.Printf("Root Command Exec:")
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
			cmd.Print("Using Google Translate")
			tr, err = models.NewGoogleTranslationRequestFromFile(args[0], source, dest, cmd)
			break
		case "gpt":
			cmd.Print("Using GPT Translate")
			tr, err = models.NewGPTTranslationRequestFromFile(args[0], source, dest, cmd)
			break
		default:
			cmd.Printf("Unknown engine %s", engine)
		}
		return actions.TranslateOne(tr)
	},
}
