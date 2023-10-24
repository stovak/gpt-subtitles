package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/asticode/go-astisub"
	"github.com/jedib0t/go-pretty/v6/table"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

type TranslationRequest interface {
	GetLogger() *zap.SugaredLogger
	GetSourceLanguage() language.Tag
	GetSourceText() []string
	GetTranslatedText() []string
	GetTargetLanguage() language.Tag
	GetTranslated() (*astisub.Subtitles, error)
	Parse() error
	String() string
	Translate() error
	WriteErrorDiff(translatedText []string) error
	WriteToFile(variant string, contents string) error
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

func (tr *TranslationRequestBase) WriteErrorDiff(translatedText []string) error {
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Source", "Translated"})
	fileName := strings.Replace(
		tr.SubtitleFileName,
		tr.Extension,
		fmt.Sprintf("_%s_error_diff.ttml", tr.TargetLanguage), 1)

	tr.GetLogger().Infof("Writing error diff to %s", fileName)
	sourceText := tr.GetSourceText()

	iterations := max(len(sourceText), len(translatedText))
	for i := 0; i < iterations; i++ {
		if len(sourceText) <= i {
			sourceText = append(sourceText, "")
		}
		if len(translatedText) <= i {
			translatedText = append(translatedText, "")
		}
		t.AppendRow(table.Row{
			sourceText[i],
			translatedText[i],
		})
	}
	return tr.WriteToFile(fmt.Sprintf("-%s-%s", "error", tr.TargetLanguage), t.Render())
}

func (tr *TranslationRequestBase) WriteToFile(variant string, contents string) error {
	fileName := strings.Replace(
		tr.SubtitleFileName,
		tr.Extension,
		fmt.Sprintf("_%s.ttml", variant), 1)
	tr.GetLogger().Infof("Writing results %s to %s", variant, fileName)
	return os.WriteFile(fileName, []byte(contents), 0700)
}
