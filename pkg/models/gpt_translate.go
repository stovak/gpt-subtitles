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
	"github.com/spf13/cobra"
	"github.com/stovak/gpt-subtitles/pkg/util"
)

type GPTTranslationRequest struct {
	TranslationRequestBase
	client          *chatgpt.Client
	results         []string
	RequestTemplate *template.Template
	SourceText      string
}

func NewGPTTranslationRequestFromFile(fileName string, sourceLanguage string, destinationLanguage string, cmd *cobra.Command) (TranslationRequest, error) {
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
			Cmd:              cmd,
		},
		SourceText:      "",
		RequestTemplate: template.Must(template.ParseFiles(path.Join(util.GetRoot(), "templates/gpt-subtitle-request.tmpl"))),
	}
	toReturn.ParseSourceTarget(sourceLanguage, destinationLanguage)
	return &toReturn, nil
}

func (tr *GPTTranslationRequest) Translate() error {
	tr.Cmd.Printf("Translating: %s %s => %s", tr.SubtitleFileName, tr.SourceLanguage, tr.TargetLanguage)
	sourceText := tr.GetSourceText()
	// Iterate over the slice in batches of 100
	for i := 0; i < len(sourceText); i += 100 {
		sourceTextSlice := sourceText[i : i+100]
		tr.Logger.Debugf("Translating %d lines", len(sourceTextSlice))
		// Call the function with the current batch of strings
		prompt, err := tr.toPrompt(sourceTextSlice)
		if err != nil {
			return err
		}
		tr.Cmd.Printf("Prompt: %s", prompt)
		req := chatgpt.ChatCompletionRequest{
			Model: chatgpt.GPT4,
			Messages: []chatgpt.ChatMessage{
				{
					Role:    chatgpt.ChatGPTModelRoleSystem,
					Content: prompt,
				},
			},
		}
		tr.Cmd.Printf("Sending a batch of %d lines to OpenAI", len(sourceText[i:i+100]))
		resp, err := tr.getClient().Send(context.Background(), &req)
		if err != nil {
			return err
		}
		if len(resp.Choices) == 0 {
			return fmt.Errorf("no choices returned")
		}
		// Split the results and then add them all to the slice of strings for results
		tr.results = append(tr.results, strings.Split(resp.Choices[0].Message.Content, "|")...)
		tr.Cmd.Printf("%d Results total", len(tr.results))

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
	tr.Cmd.Printf("Prompt: %s => %s", buf.String(), err)
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
