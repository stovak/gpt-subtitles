package actions

import "github.com/stovak/gpt-subtitles/pkg/models"

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
	err = tr.WriteTranslatedToNewFile()
	if err != nil {
		tr.GetLogger().Errorf("%s => %s:Error writing translated file: %s", tr.GetSourceLanguage(), tr.GetTargetLanguage(), err)
	}
	return err
}
