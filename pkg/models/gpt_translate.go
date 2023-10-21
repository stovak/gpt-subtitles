package models

import (
	"context"
	"fmt"
	"github.com/stovak/gpt-subtitles/pkg/util"
	"html/template"
	"log"
	"os"
	"path"
	"strings"

	"github.com/asticode/go-astisub"
	"github.com/ayush6624/go-chatgpt"
	"go.uber.org/zap"
)

type GPTTranslationRequest struct {
	TranslationRequestBase
	client          *chatgpt.Client
	results         *chatgpt.ChatResponse
	RequestTemplate *template.Template
	SourceText      string
}

func NewGPTTranslationRequestFromFile(fileName string, sourceLanguage string, destinationLanguage string, log *zap.SugaredLogger) (TranslationRequest, error) {
	subs, err := astisub.OpenFile(fileName)
	if err != nil {
		return &GPTTranslationRequest{}, err
	}
	if err != nil {
		return &GPTTranslationRequest{}, err
	}
	toReturn := GPTTranslationRequest{
		TranslationRequestBase: TranslationRequestBase{
			SubtitleFileName: fileName,
			Subtitles:        subs,
			Logger:           log,
		},
	}
	toReturn.ParseSourceTarget(sourceLanguage, destinationLanguage)
	return &toReturn, nil
}

func (tr *GPTTranslationRequest) Translate() error {
	var err error
	tr.Logger.Debugf("Translating: %s %s => %s", tr.SubtitleFileName, tr.SourceLanguage, tr.TargetLanguage)
	prompt, err := tr.toPrompt()
	if err != nil {
		return err
	}
	tr.Logger.Debugf("Prompt: %s", prompt)
	req := chatgpt.ChatCompletionRequest{
		Model: chatgpt.GPT4,
		Messages: []chatgpt.ChatMessage{
			{
				Role:    chatgpt.ChatGPTModelRoleSystem,
				Content: prompt,
			},
		},
	}
	//tr.Logger.Debugf("Sending request: %#v", req)
	tr.results, err = tr.getClient().Send(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (tr *GPTTranslationRequest) toPrompt() (string, error) {
	tr.RequestTemplate = template.Must(template.ParseFiles(path.Join(util.GetRoot(), "templates/gpt-subtitle-request.tmpl")))
	var err error
	var buf = new(strings.Builder)

	// Make sure input has been parsed
	if tr.Subtitles == nil {
		tr.Subtitles, err = astisub.OpenFile(tr.SubtitleFileName)
		if err != nil {
			return "", err
		}
	}
	tr.SourceText = strings.Join(tr.GetSourceText(), "|")
	err = tr.RequestTemplate.Execute(buf, tr)
	tr.Logger.Debugf("Prompt: %s => %s", buf.String(), err)
	return buf.String(), err
}

func (tr *GPTTranslationRequest) getClient() *chatgpt.Client {
	if tr.client == nil {
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			log.Fatal("OPENAI_API_KEY environment variable not set")
		}
		var err error
		tr.client, err = chatgpt.NewClient(key)
		if err != nil {
			log.Fatal(err)
		}
	}
	return tr.client
}

// GetTranslated returns a new Subtitles object with the translated text
// err is non-nil if there was an error translating
func (tr *GPTTranslationRequest) GetTranslated() (*astisub.Subtitles, error) {
	var err error
	if tr.results == nil {
		return nil, fmt.Errorf("no results to translate")
	}
	toReturn := astisub.NewSubtitles()
	r, _ := tr.TargetLanguage.Region()
	region := &astisub.Region{
		ID: r.String(),
	}

	toReturn.Regions = map[string]*astisub.Region{
		region.ID: region,
	}
	if err != nil {
		return nil, err
	}
	resultLines := strings.Split(tr.results.Choices[0].Message.Content, "|")
	if len(resultLines) != len(tr.Subtitles.Items) {
		tr.Logger.Errorf("number of lines in result (%d) does not match number of lines in source (%d)", len(resultLines), len(tr.Subtitles.Items))
		tr.Logger.Errorf("Returned Translation: %#v", tr.results)
		return nil, fmt.Errorf("number of lines in result (%d) does not match number of lines in source (%d)", len(resultLines), len(tr.Subtitles.Items))
	}
	for num, item := range tr.Subtitles.Items {
		toReturn.Items = append(toReturn.Items, &astisub.Item{
			Region:  region,
			StartAt: item.StartAt,
			EndAt:   item.EndAt,
			Lines: []astisub.Line{
				{
					Items: []astisub.LineItem{
						{
							Text: resultLines[num],
						},
					},
				},
			},
		})
	}
	return toReturn, nil
}

func (tr *GPTTranslationRequest) WriteTranslatedToNewFile() error {
	fileName := strings.Replace(
		tr.SubtitleFileName,
		tr.Extension,
		fmt.Sprintf("_%s.ttml", tr.TargetLanguage), 1)

	tr.GetLogger().Infof("Writing results to %s", fileName)
	translated, err := tr.GetTranslated()
	if err != nil {
		return err
	}
	return translated.Write(fileName)
}
