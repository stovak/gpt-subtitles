package models

import (
	"context"
	"github.com/asticode/go-astisub"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/api/translate"
)

type GoogleTranslateRequest struct {
	TranslationRequestBase
	client  *translate.Client
	results []translate.Translation
}

func NewGoogleTranslationRequestFromFile(fileName string, sourceLanguage string, destinationLanguage string, log *zap.SugaredLogger) (TranslationRequest, error) {
	subs, err := astisub.OpenFile(fileName)
	if err != nil {
		return &GoogleTranslateRequest{}, err
	}
	toReturn := GoogleTranslateRequest{
		TranslationRequestBase: TranslationRequestBase{
			SubtitleFileName: fileName,
			Subtitles:        subs,
			Logger:           log,
		},
	}
	toReturn.ParseSourceTarget(sourceLanguage, destinationLanguage)
	toReturn.Extension = filepath.Ext(fileName)
	return &toReturn, nil
}

func (tr *GoogleTranslateRequest) Translate() error {
	var err error
	tr.Logger.Infof("Translating %s to %s", tr.SourceLanguage, tr.TargetLanguage)
	sourceText := tr.GetSourceText()
	tr.Logger.Infof("Source Text: %#v", sourceText)
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
			tr.Logger.Fatal("GOOGLE_TRANSLATE_API_KEY environment variable not set")
		}
		var err error
		tr.client, err = translate.NewClient(context.Background(), option.WithCredentialsFile(os.ExpandEnv("$HOME/.keys/subtitles-translator@dog-park-adjacent.iam.gserviceaccount.com.key")))
		if err != nil {
			tr.Logger.Fatal(err)
		}
	}
	return tr.client
}
