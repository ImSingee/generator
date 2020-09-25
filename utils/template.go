package utils

import (
	"log"
	"text/template"
)

func GetTemplate(name string, text string) *template.Template {
	t, err := template.New(name).Parse(text)

	if err != nil {
		log.Fatalln(err)
	}

	return t
}
