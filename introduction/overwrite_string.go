package introduction

import (
	"strings"
)

// OverwriteString overwrites the first 'count' characters in a string with
// the rune 'value'
func OverwriteString(str string, value rune, count int) string {
	// If asked for more no need to loop, just return string length * the rune
	if count > len(str) {
		return strings.Repeat(string(value), len(str))
	}

	result := []rune(str)
	for i := 0; i <= count; i++ {
		result[i] = value
	}
	return string(result)
}
