package helpers

import "strings"

func NextBridgeID(c string, maxLen int) string {
	if c == "" {
		return "a"
	}
	if len(c) == maxLen && c == strings.Repeat("z", maxLen) {
		return ""
	}

	ca := []rune(c)
	i := len(ca) - 1
	for i >= 0 && ca[i] == 'z' {
		ca[i] = 'a'
		i--
	}
	if i < 0 {
		ca = append([]rune{'a'}, ca...)
	} else {
		ca[i]++
	}
	return string(ca)
}
