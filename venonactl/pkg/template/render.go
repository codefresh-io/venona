package template

import (
	"bytes"
	"html/template"

	"github.com/sirupsen/logrus"
)

type (
	TemplateRender interface {
		Render(string, interface{}) (string, error)
	}

	render struct {
		logger logrus.Logger
	}
)

func NewRender() TemplateRender {
	return &render{}
}

func (r *render) Render(tmpl string, data interface{}) (string, error) {
	template, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBufferString("")
	err = template.Execute(buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
