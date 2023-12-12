package controllers

import (
	"html/template"
	"net/http"

	"github.com/taherk/galleryapp/views"
)

func StaticHandler(tpl views.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, nil)
	}
}

func FAQ(tpl views.Template) http.HandlerFunc {
	faqs := []struct {
		Question template.HTML
		Answer   template.HTML
	}{
		{
			"Is there a free version?",
			"Yes! We offer a free trial for 30 days on any paid plans.",
		},
		{
			"What are your support hours?",
			"We have support staff answering emails 24/7, though response times may be a bit slower on weekends.",
		},
		{
			"How do I contact support?",
			`Email us - <a href="mailto:support@kathanawala.com">support@kathanawala.com</a>`,
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, faqs)
	}
}
