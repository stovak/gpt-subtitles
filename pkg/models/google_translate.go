package models

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/translate"
	"github.com/asticode/go-astisub"
	"google.golang.org/api/option"
)

type GoogleTranslateRequest struct {
	TranslationRequestBase
	client  *translate.Client
	results []translate.Translation
}

func NewGoogleTranslationRequestFromFile(fileName string, sourceLanguage string, destinationLanguage string, cmd *cobra.Command) (TranslationRequest, error) {
	subs, err := astisub.OpenFile(fileName)
	if err != nil {
		return &GoogleTranslateRequest{}, err
	}
	toReturn := GoogleTranslateRequest{
		TranslationRequestBase: TranslationRequestBase{
			SubtitleFileName: fileName,
			Subtitles:        subs,
			Cmd:              cmd,
		},
	}
	toReturn.ParseSourceTarget(sourceLanguage, destinationLanguage)
	toReturn.Extension = filepath.Ext(fileName)
	return &toReturn, nil
}

func (tr *GoogleTranslateRequest) Translate() error {
	var err error
	tr.Cmd.Printf("Translating %s to %s", tr.SourceLanguage, tr.TargetLanguage)
	sourceText := tr.GetSourceText()
	tr.Cmd.Printf("Source Text: %#v", sourceText)
	tr.results, err = tr.getClient().Translate(context.Background(), tr.GetSourceText(), tr.TargetLanguage, &translate.Options{
		Source: tr.SourceLanguage,
		Format: translate.Text,
		Model:  "nmt",
	})
	return err
}

func (tr *GoogleTranslateRequest) GetTranslated() (*astisub.Subtitles, error) {
	var err error
	if tr.results == nil {
		err = tr.Translate()
		if err != nil {
			return nil, err
		}
	}
	toReturn := astisub.NewSubtitles()
	for i, result := range tr.results {
		toReturn.Items = append(toReturn.Items, &astisub.Item{
			StartAt: tr.Subtitles.Items[i].StartAt,
			EndAt:   tr.Subtitles.Items[i].EndAt,
			Lines: []astisub.Line{
				{
					Items: []astisub.LineItem{
						{
							Text: result.Text,
						},
					},
				},
			},
		})
	}
	return toReturn, nil
}

func (tr *GoogleTranslateRequest) getClient() *translate.Client {
	if tr.client == nil {
		key := os.Getenv("GOOGLE_TRANSLATE_API_KEY")
		if key == "" {
			tr.Cmd.Printf("GOOGLE_TRANSLATE_API_KEY environment variable not set")
		}
		var err error
		tr.client, err = translate.NewClient(context.Background(), option.WithCredentialsFile(os.ExpandEnv("$HOME/.keys/subtitles-translator@dog-park-adjacent.iam.gserviceaccount.com.key")))
		if err != nil {
			tr.Cmd.PrintErrf("Translate get client error: %s", err)
		}
	}
	return tr.client
}

func (tr *GoogleTranslateRequest) WriteTranslatedToNewFile() error {
	fileName := strings.Replace(
		tr.SubtitleFileName,
		tr.Extension,
		fmt.Sprintf("_%s.ttml", tr.TargetLanguage), 1)

	log.Printf("Writing results to %s", fileName)
	translated, err := tr.GetTranslated()
	if err != nil {
		return err
	}
	return translated.Write(fileName)
}

func (tr *GoogleTranslateRequest) GetTranslatedText() []string {
	var toReturn []string
	for _, result := range tr.results {
		toReturn = append(toReturn, result.Text)
	}
	return toReturn
}
