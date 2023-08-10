package main

import (
	"github.com/chauvinhphuoc/snippetbox/internal/db/sqlc"
	"time"
)

// templateData acts as the holding structure for any dynamic data
// that we want to pass to out HTML templates.
type templateData struct {
	CurrentYear int
	Snippet     sqlc.Snippet
	Snippets    []sqlc.Snippet
}

// newTemplateData returns a *templateData, which contains some fields having default values.
func (app *application) newTemplateData() *templateData {
	return &templateData{
		CurrentYear: time.Now().Year(),
	}
}
