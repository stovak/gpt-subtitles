package models

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"

	"github.com/asticode/go-astisub"
	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/text/language"
)

type TranslationRequest interface {
	GetSourceLanguage() language.Tag
	GetSourceText() []string
	GetCmd() *cobra.Command
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
	Cmd              *cobra.Command
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
	tr.Cmd.Printf("Split text: %+v", toReturn)
	return toReturn
}

func (tr *TranslationRequestBase) String() string {
	buf := new(strings.Builder)
	err := tr.Subtitles.WriteToTTML(buf)
	if err != nil {
		tr.Cmd.Printf("Error writing to string: %s", err)
	}
	return buf.String()
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

	tr.Cmd.Printf("Writing error diff to %s", fileName)
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
	tr.Cmd.Printf("Writing results %s to %s", variant, fileName)
	return os.WriteFile(fileName, []byte(contents), 0700)
}

func (tr *TranslationRequestBase) GetCmd() *cobra.Command {
	return tr.Cmd
}
