package models

import (
	"html/template"
	"io"
	"os"
	"strings"
)

type TranslationRequest struct {
	SubtitleFileName     string
	Extension            string
	SourceLanguage       string
	DestinationLanguage  string
	SubtitleFileContents string
}

func NewTranslationRequestFromFile(fileName string, sourceLanguage string, destinationLanguage string) TranslationRequest {
	toReturn := TranslationRequest{
		SubtitleFileName:     fileName,
		SourceLanguage:       sourceLanguage,
		DestinationLanguage:  destinationLanguage,
		SubtitleFileContents: "",
	}
	explodedName := strings.Split(fileName, ".")
	toReturn.Extension = explodedName[len(explodedName)-1]

	return toReturn
}

func NewTranslationRequestFromContents(contents string, extension string, sourceLanguage string, destinationLanguage string) TranslationRequest {
	return TranslationRequest{
		SubtitleFileContents: contents,
		Extension:            extension,
		SourceLanguage:       sourceLanguage,
		DestinationLanguage:  destinationLanguage,
	}
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
