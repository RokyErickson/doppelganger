package prompt

import (
	"strings"
)

type PromptKind uint8

const (
	PromptKindSecret PromptKind = iota
	PromptKindEcho
	PromptKindBinary
)

var binaryPromptSuffixes = []string{
	"(yes/no)? ",
	"(yes/no): ",
}

func Classify(prompt string) PromptKind {
	for _, suffix := range binaryPromptSuffixes {
		if strings.HasSuffix(prompt, suffix) {
			return PromptKindBinary
		}
	}
	return PromptKindSecret
}
