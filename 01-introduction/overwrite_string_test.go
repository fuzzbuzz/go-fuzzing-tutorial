package introduction

import (
	"testing"
	"unicode/utf8"
)

func FuzzBasicOverwriteString(f *testing.F) {
	f.Add("Hello, world!", 'A', 15)

	f.Fuzz(func(t *testing.T, str string, value rune, count int) {
		OverwriteString(str, value, count)
	})
}

// nah
func FuzzOverwriteStringSuffix(f *testing.F) {
	f.Add("Hello, world!", 'A', 15)

	f.Fuzz(func(t *testing.T, str string, value rune, count int) {
		result := OverwriteString(str, value, count)
		if count > 0 && count < utf8.RuneCountInString(str) {
			resultSuffix := string([]rune(result)[count:])
			strSuffix := string([]rune(str)[count:])
			if resultSuffix != strSuffix {
				t.Fatalf("OverwriteString modified too many characters! Expected %s, got %s.", strSuffix, resultSuffix)
			}
		}
	})
}
