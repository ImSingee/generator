package utils

import (
	"bytes"
	"log"
	"text/template"
)

func GetTemplate(name string, text string) *template.Template {
	t, err := template.New(name).Parse(text)

	if err != nil {
		log.Fatalf("Cannot build template: name = %s; text = %s; err = %s", name, text, err)
	}

	return t
}

func ExecuteTemplate(tmpl *template.Template, data interface{}) string {
	b := bytes.NewBuffer(make([]byte, 0, 64))

	err := tmpl.Execute(b, data)

	if err != nil {
		log.Fatalf("Cannot execute template: template = %#+v, err = %s", tmpl, err)
	}

	return b.String()
}
