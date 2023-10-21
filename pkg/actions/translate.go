package actions

import (
	"github.com/stovak/gpt-subtitles/pkg/models"
	"strings"
)

func TranslateOne(tr models.TranslationRequest) error {
	tr.GetLogger().Infof("Translating %s => %s", tr.GetSourceLanguage(), tr.GetTargetLanguage())
	err := tr.Parse()
	if err != nil {
		tr.GetLogger().Errorf("%s => %s:Error parsing file: %s", tr.GetSourceLanguage(), tr.GetTargetLanguage(), err)
		return err
	}
	err = tr.Translate()
	if err != nil {
		tr.GetLogger().Errorf("%s => %s:Error translating file: %s", tr.GetSourceLanguage(), tr.GetTargetLanguage(), err)
		return err
	}
	translated, err := tr.GetTranslated()
	if err != nil {
		tr.GetLogger().Errorf("%s => %s:Error getting translated file: %s", tr.GetSourceLanguage(), tr.GetTargetLanguage(), err)
		return tr.WriteErrorDiff(tr.GetTranslatedText())
	}
	buff := new(strings.Builder)
	err = translated.WriteToTTML(buff)
	if err != nil {
		tr.GetLogger().Errorf("%s => %s:Error writing translated file: %s", tr.GetSourceLanguage(), tr.GetTargetLanguage(), err)
		return tr.WriteErrorDiff(tr.GetTranslatedText())
	}
	err = tr.WriteToFile(tr.GetTargetLanguage().String(), buff.String())
	if err != nil {
		tr.GetLogger().Errorf("%s => %s:Error writing translated file: %s", tr.GetSourceLanguage(), tr.GetTargetLanguage(), err)
	}
	return err
}
