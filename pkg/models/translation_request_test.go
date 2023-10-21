package models

import (
	"fmt"
	"github.com/asticode/go-astisub"
	"github.com/stovak/gpt-subtitles/pkg/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"golang.org/x/text/language"
	"path"
	"testing"
)

var (
	observedZapCore, observedLogs = observer.New(zap.InfoLevel)
	observedLogger                = zap.New(observedZapCore)
)

func TestTranslationRequestBase_ParseSourceTarget(t *testing.T) {

	type args struct {
		source string
		target string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ParseSourceTarget-1",
			args: args{
				source: "en",
				target: "es",
			},
		},
		{
			name: "ParseSourceTarget-2",
			args: args{
				source: "jv",
				target: "en",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TranslationRequestBase{
				Logger: observedLogger.Sugar(),
			}
			tr.ParseSourceTarget(tt.args.source, tt.args.target)
			assert.IsTypef(t, language.Tag{}, tr.SourceLanguage, "SourceLanguage is not a language.Tag")
			base, _ := tr.SourceLanguage.Base()
			assert.Equalf(t, tt.args.source, base.String(), "Error: Source languages do not match %s => %s", tt.args.source, base)
			assert.IsTypef(t, language.Tag{}, tr.TargetLanguage, "TargetLanguage is not a language.Tag")
			base, _ = tr.TargetLanguage.Base()
			assert.Equalf(t, tt.args.target, base.String(), "Error: Target languages do not match %s => %s", tt.args.target, base)
		})
	}
}

func TestTranslationRequestBase_Parse(t *testing.T) {
	type fields struct {
		SubtitleFileName string
		Extension        string
		Subtitles        *astisub.Subtitles
		Logger           *zap.SugaredLogger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Parse-1",
			fields: fields{
				SubtitleFileName: "TestFixture1.ttml",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TranslationRequestBase{
				SubtitleFileName: path.Join(util.GetRoot(), "test-fixtures", tt.fields.SubtitleFileName),
				Extension:        tt.fields.Extension,
				Subtitles:        tt.fields.Subtitles,
				Logger:           tt.fields.Logger,
			}
			tt.wantErr(t, tr.Parse(), fmt.Sprintf("Parse()"))
			assert.Equalf(t, ".ttml", tr.Extension, "Extension is not .ttml")
			assert.IsTypef(t, &astisub.Subtitles{}, tr.Subtitles, "Subtitles is not an astisub.Subtitles")
			assert.NotZerof(t, len(tr.Subtitles.Items), "Subtitles.Items is empty")
		})
	}
}
