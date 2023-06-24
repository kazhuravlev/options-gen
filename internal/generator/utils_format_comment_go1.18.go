//go:build go1.18
// +build go1.18

package generator

import (
	"bytes"
	"strings"
)

func prefix(str1, str2 string) string {
	if str1 == "" || str2 == "" {
		return ""
	}

	buf := bytes.NewBuffer(nil)
	if len(str1) > len(str2) {
		str1, str2 = str2, str1
	}

	for i := range str1 {
		if str1[i] != str2[i] {
			break
		}

		buf.WriteByte(str1[i])
	}

	return buf.String()
}

// formatComment is a hacked version for go1.18 which has another comment format.
func formatComment(comment string) string {
	if comment == "" {
		return ""
	}

	buf := bytes.NewBuffer(nil)

	lines := strings.Split(comment, "\n")
	commonPrefix := lines[0]
	for _, line := range lines {
		commonPrefix = prefix(commonPrefix, line)
	}

	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], commonPrefix)

		// Last line contains an empty string.
		if lines[i] == "" && i == len(lines)-1 {
			continue
		}

		if i != 0 {
			buf.WriteString("\n")
		}

		buf.WriteString("// ")
		buf.WriteString(lines[i])
	}

	return buf.String()
}
