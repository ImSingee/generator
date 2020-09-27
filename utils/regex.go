package utils

import (
	"log"
	"regexp"
)

func CompileRegex(pattern string) *regexp.Regexp {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatalf("Cannot compile regex %s: %s", pattern, err)
	}

	return regex
}
