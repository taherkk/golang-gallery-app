package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/taherk/galleryapp/context"
	"github.com/taherk/galleryapp/models"
)

func Must(tpl Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return tpl
}

func ParseFS(fs fs.FS, patterns ...string) (Template, error) {
	tpl := template.New(patterns[0])
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() (template.HTML, error) {
				return "", fmt.Errorf("csrf field to be replaced")
			},
			"currentUser": func() (template.HTML, error) {
				return "", fmt.Errorf("current user not implemented")
			},
		},
	)

	htmlTpl, err := tpl.ParseFS(fs, patterns...)
	if err != nil {
		return Template{}, fmt.Errorf("parsing template: %w", err)
	}
	return Template{
		htmlTpl: htmlTpl,
	}, nil
}

//func Parse(filepath string) (Template, error) {
//	htmlTpl, err := template.ParseFiles(filepath)
//	if err != nil {
//		return Template{}, fmt.Errorf("parsing template: %w", err)
//	}
//	return Template{
//		htmlTpl: htmlTpl,
//	}, nil
//}

type Template struct {
	htmlTpl *template.Template
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data interface{}) {
	// If there multiple web reqs then all will be pointing to the same
	// template and if two requests come simultaneously then they will
	// share the token which is not good
	tpl, err := t.htmlTpl.Clone()
	if err != nil {
		fmt.Printf("cloning template: %v", err)
		http.Error(w, "There was an error executing the template", http.StatusInternalServerError)
	}
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() template.HTML {
				return csrf.TemplateField(r)
			},
			"currentUser": func() *models.User {
				return context.User(r.Context())
			},
		},
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	// if there is an error you will write to the header twice
	// which will cause another error. So you could use bytes.Buffer
	// to make sure there is no error in executing the template
	// Additionally html template will be written to w till point
	// the error occurs resulting in half rendered page.
	err = tpl.Execute(&buf, data)
	if err != nil {
		log.Printf("executing template: %v", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf)
}
