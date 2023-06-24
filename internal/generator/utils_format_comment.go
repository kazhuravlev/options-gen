//go:build !go1.18
// +build !go1.18

package generator

import (
	"bytes"
	"strings"
)

func formatComment(comment string) string {
	if comment == "" {
		return ""
	}

	buf := bytes.NewBuffer(nil)

	lines := strings.Split(comment, "\n")
	for i := range lines {
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
