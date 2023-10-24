package models

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"strings"

	"github.com/asticode/go-astisub"
	"github.com/ayush6624/go-chatgpt"
	"github.com/stovak/gpt-subtitles/pkg/util"
	"go.uber.org/zap"
)

type GPTTranslationRequest struct {
	TranslationRequestBase
	client          *chatgpt.Client
	results         []string
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
			Extension:        path.Ext(fileName),
			Subtitles:        subs,
			Logger:           log,
		},
		SourceText:      "",
		RequestTemplate: template.Must(template.ParseFiles(path.Join(util.GetRoot(), "templates/gpt-subtitle-request.tmpl"))),
	}
	toReturn.ParseSourceTarget(sourceLanguage, destinationLanguage)
	return &toReturn, nil
}

func (tr *GPTTranslationRequest) Translate() error {
	tr.Logger.Debugf("Translating: %s %s => %s", tr.SubtitleFileName, tr.SourceLanguage, tr.TargetLanguage)
	sourceText := tr.GetSourceText()
	// Iterate over the slice in batches of 100
	for i := 0; i < len(sourceText); i += 100 {
		sourceTextSlice := sourceText[i : i+100]
		sourceTextSlice = util.TrimSlice(sourceTextSlice)
		// Call the function with the current batch of strings
		prompt, err := tr.toPrompt(sourceTextSlice)
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
		tr.Logger.Infof("Sending a batch of %d lines to OpenAI", len(sourceTextSlice))
		resp, err := tr.getClient().Send(context.Background(), &req)
		if err != nil || len(resp.Choices) == 0 {
			return err
		}
		translatedText := strings.Split(resp.Choices[0].Message.Content, "|")
		diff := len(sourceText) - len(translatedText)
		if diff != 0 {
			tr.Logger.Warnf("Translated text length (%d) does not match source text length (%d)", len(translatedText), len(sourceTextSlice))
		}
		for i := 0; i < diff; i++ {
			// Add blank spaces and deal with with this later
			translatedText = append(translatedText, " ")
		}

		// Split the results and then add them all to the slice of strings for results
		tr.results = append(tr.results, translatedText...)
		tr.GetLogger().Debugf("%d Results total", len(tr.results))
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func (tr *GPTTranslationRequest) toPrompt(batch []string) (string, error) {
	var err error
	var buf = new(strings.Builder)
	// String the source text together in a single string separated by |
	tr.SourceText = strings.Join(batch, "|")
	// Execute the template and capture the output
	err = tr.RequestTemplate.Execute(buf, tr)
	if err != nil {
		return "", err
	}
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
	if len(tr.results) != len(tr.Subtitles.Items) {
		_ = tr.WriteErrorDiff(tr.results)
		return nil, fmt.Errorf("number of lines in result (%d) does not match number of lines in source (%d)", len(tr.results), len(tr.Subtitles.Items))
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
							Text: tr.results[num],
						},
					},
				},
			},
		})
	}
	return toReturn, nil
}

func (tr *GPTTranslationRequest) GetTranslatedText() []string {
	return tr.results
}
