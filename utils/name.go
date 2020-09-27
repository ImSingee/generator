package utils

import (
	"fmt"
	"strings"
	"unicode"
)

func GetShortName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name cannot be empty")
	}

	shortName := strings.Builder{}

	for _, c := range name {
		if unicode.IsUpper(c) {
			shortName.WriteRune(unicode.ToLower(c))
		}
	}

	if shortName.Len() == 0 {
		for _, c := range name {
			if unicode.IsUpper(c) || unicode.IsLower(c) {
				shortName.WriteRune(unicode.ToLower(c))
				break
			}
		}
	}

	if shortName.Len() == 0 {
		return "", fmt.Errorf("invalid name: %s", name)
	}

	return shortName.String(), nil
}

func IsUpper(c byte) bool {
	return c >= byte('A') && c <= byte('Z')
}

func IsLower(c byte) bool {
	return c >= byte('a') && c <= byte('z')
}

func IsASCII(c byte) bool {
	return c&0b1000_0000 == 0
}

// public 的要求
func IsPublic(name string) bool {
	c := name[0]

	return IsASCII(c) && IsUpper(c)
}

func IsPrivate(name string) bool {
	return !IsPublic(name)
}

func ShouldIgnore(name string) bool {
	return name == "" || !IsASCII(name[0])
}

func ToGetterName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name cannot be empty")
	}
	if !IsASCII(name[0]) {
		return "", fmt.Errorf("name %s must started with an ASCII letter", name)
	}
	if IsUpper(name[0]) {
		return "", fmt.Errorf("name %s must started with a lower ASCII letter", name)
	}

	return toGetterName(name), nil
}

func toGetterName(name string) string {
	getterName := strings.Builder{}

	n := []rune(name)

	getterName.WriteRune(unicode.ToUpper(n[0]))

	for _, c := range n[1:] {
		getterName.WriteRune(c)
	}

	return getterName.String()
}
