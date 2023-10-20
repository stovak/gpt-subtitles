package models

import (
	"github.com/asticode/go-astisub"
	"go.uber.org/zap"
	"html/template"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ayush6624/go-chatgpt"
)

type TranslationRequest struct {
	SubtitleFileName     string
	Extension            string
	SourceLanguage       string
	DestinationLanguage  string
	SubtitleFileContents string
	Logger               *zap.SugaredLogger
}

func NewTranslationRequestFromFile(fileName string, sourceLanguage string, destinationLanguage string, log *zap.SugaredLogger) (TranslationRequest, error) {
	subs, err := astisub.OpenFile(fileName)
	if err != nil {
		return TranslationRequest{}, err
	}
	buf := new(strings.Builder)
	err = subs.WriteToSRT(buf)
	toReturn := TranslationRequest{
		SubtitleFileName:     fileName,
		SourceLanguage:       Languages[sourceLanguage],
		DestinationLanguage:  Languages[destinationLanguage],
		SubtitleFileContents: buf.String(),
		Logger:               log,
	}
	explodedName := strings.Split(fileName, ".")
	toReturn.Extension = explodedName[len(explodedName)-1]

	return toReturn, nil
}

func NewTranslationRequestFromContents(contents string, extension string, sourceLanguage string, destinationLanguage string, log *zap.SugaredLogger) TranslationRequest {
	return TranslationRequest{
		SubtitleFileContents: contents,
		Extension:            extension,
		SourceLanguage:       sourceLanguage,
		DestinationLanguage:  destinationLanguage,
		Logger:               log,
	}
}

func (tr *TranslationRequest) String() string {
	buf := new(strings.Builder)
	err := tr.ToPrompt(buf)
	if err != nil {
		return ""
	}
	return buf.String()
}

func (tr *TranslationRequest) ToPrompt(w io.Writer) error {
	var t = template.Must(template.ParseFiles("templates/gpt-subtitle-request.tmpl"))

	if tr.SubtitleFileContents == "" {
		fileContents, err := os.ReadFile(tr.SubtitleFileName)
		if err != nil {
			panic(err)
		}
		tr.SubtitleFileContents = string(fileContents)
	}

	return t.Execute(w, tr)
}

func (tr *TranslationRequest) WriteResultsToFile(results *chatgpt.ChatResponse) {
	fileName := strings.Replace(tr.SubtitleFileName, tr.SourceLanguage, tr.DestinationLanguage, 1)
	log.Printf("Writing results to %s", fileName)
	splitText := strings.Split(results.Choices[0].Message.Content, "===")
	tr.Logger.Debugf("Split text: %+v", splitText)
	//err := os.WriteFile(fileName, []byte(), 0700)
	//if err != nil {
	//	panic(err)
	//}
}
