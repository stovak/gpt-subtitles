package models

import (
	"context"
	"fmt"
	"github.com/stovak/gpt-subtitles/pkg/util"
	"html/template"
	"log"
	"os"
	"path"
	"reflect"
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
	tr.Logger.Debugf("Translating: %s", tr.SubtitleFileName)
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
	tr.Logger.Debugf("Sending request: %#v", req)
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
	tr.SourceText = strings.Join(tr.GetSourceText(), "\n")
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
		err := tr.Translate()
		if err != nil {
			return nil, err
		}
	}
	toReturn := reflect.ValueOf(tr.Subtitles).Elem().Interface().(astisub.Subtitles)
	region, _ := tr.TargetLanguage.Region()

	toReturn.Regions = map[string]*astisub.Region{
		region.String(): &astisub.Region{
			ID: region.String(),
		},
	}
	if err != nil {
		return nil, err
	}
	resultLines := strings.Split(tr.results.Choices[0].Message.Content, "\n")
	for i, _ := range tr.Subtitles.Items {
		toReturn.Items[i].Lines[0].Items[0].Text = resultLines[i]
	}
	return &toReturn, nil
}

func (tr *GPTTranslationRequest) WriteTranslatedToNewFile() error {
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
