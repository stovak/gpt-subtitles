package models

import (
	"path/filepath"
	"strings"

	"github.com/asticode/go-astisub"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

type TranslationRequest interface {
	Translate() error
	WriteTranslatedToNewFile() error
	Parse() error
	String() string
	GetTranslated() (*astisub.Subtitles, error)
	GetSourceText() []string
	GetLogger() *zap.SugaredLogger
	GetSourceLanguage() language.Tag
	GetTargetLanguage() language.Tag
}

type TranslationRequestBase struct {
	SourceLanguage   language.Tag
	TargetLanguage   language.Tag
	SubtitleFileName string
	Extension        string
	Subtitles        *astisub.Subtitles
	Logger           *zap.SugaredLogger
}

func (tr *TranslationRequestBase) ParseSourceTarget(source string, target string) {
	tr.SourceLanguage = language.MustParse(source)
	tr.TargetLanguage = language.MustParse(target)
}

func (tr *TranslationRequestBase) Parse() error {
	var err error
	tr.Subtitles, err = astisub.OpenFile(tr.SubtitleFileName)
	tr.Extension = filepath.Ext(tr.SubtitleFileName)
	return err
}

// GetSourceText returns a string of all the lines in the subtitle file
func (tr *TranslationRequestBase) GetSourceText() []string {
	var toReturn []string
	for _, item := range tr.Subtitles.Items {
		for _, line := range item.Lines {
			toReturn = append(toReturn, line.String())
		}
	}
	tr.Logger.Debugf("Split text: %+v", toReturn)
	return toReturn
}

func (tr *TranslationRequestBase) String() string {
	buf := new(strings.Builder)
	err := tr.Subtitles.WriteToTTML(buf)
	if err != nil {
		tr.Logger.Fatalf("Error writing to string: %s", err)
	}
	return buf.String()
}

func (tr *TranslationRequestBase) GetLogger() *zap.SugaredLogger {
	return tr.Logger
}

func (tr *TranslationRequestBase) GetSourceLanguage() language.Tag {
	return tr.SourceLanguage
}

func (tr *TranslationRequestBase) GetTargetLanguage() language.Tag {
	return tr.TargetLanguage
}
