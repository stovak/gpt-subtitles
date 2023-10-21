package models

import (
	"fmt"
	"path"
	"testing"

	"github.com/asticode/go-astisub"
	"github.com/stovak/gpt-subtitles/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestGoogleTranslateRequest_Translate(t *testing.T) {

	tests := []struct {
		name                string
		fileName            string
		sourceLanguage      string
		destinationLanguage string
		wantErr             assert.ErrorAssertionFunc
	}{
		{
			name:                "Translate-1",
			fileName:            "TestFixture1.ttml",
			sourceLanguage:      "en",
			destinationLanguage: "es",
			wantErr:             assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := NewGoogleTranslationRequestFromFile(
				path.Join(util.GetRoot(), "test-fixtures", tt.fileName),
				tt.sourceLanguage, tt.destinationLanguage, observedLogger.Sugar())
			assert.NoError(t, err, fmt.Sprintf("NewGoogleTranslationRequestFromFile(%s, %s, %s)", tt.fileName, tt.sourceLanguage, tt.destinationLanguage))
			t.Logf("tr: %#v", tr)
			st := tr.GetSourceText()
			t.Logf("Source Text: %#v", st)
			err = tr.Translate()
			tt.wantErr(t, err, fmt.Sprintf("Translate Error: %s %#v", err, observedLogs.All()))
			tlated, err := tr.GetTranslated()
			assert.NoError(t, err, fmt.Sprintf("GetTranslated()"))
			assert.IsTypef(t, astisub.Subtitles{}, *tlated, "results is not a subtitles object")
		})
	}
}
