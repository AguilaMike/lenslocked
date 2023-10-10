package controllers

import (
	"html/template"

	"github.com/go-chi/chi/v5"
)

func FAQ(r *chi.Mux) {
	questions := []struct {
		Question string
		Answer   template.HTML
	}{
		{
			Question: "Is there a free version?",
			Answer:   "Yes! We offer a free trial for 30 days on any paid plans.",
		},
		{
			Question: "What are your support hours?",
			Answer:   "We have support staff answering emails 24/7, though response times may be a bit slower on weekends.",
		},
		{
			Question: "How do I contact support?",
			Answer:   `Email us - <a href="mailto:support@test.com">support@test.com</a>`,
		},
		{
			Question: "Where is your office?",
			Answer:   "Our entire team is remote!",
		},
	}

	RegisterGetControllerWithTemplateFs(r, "/faq", "faq.gohtml", questions)
}
