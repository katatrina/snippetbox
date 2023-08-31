package main

import (
	"github.com/chauvinhphuoc/snippetbox/internal/db/sqlc"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

// functionTemplates contains all baked-in functions which integrated in every template set.
var functionTemplates = template.FuncMap{
	"humanDate": humanDate,
}

// templateData acts as the holding structure for any dynamic data
// that we want to pass to our HTML templates.
type templateData struct {
	CurrentYear     int                 // used for printing current year
	Snippet         sqlc.Snippet        // used for view snippet page
	Snippets        []sqlc.Snippet      // used for home page
	Form            any                 // used for any HTML form
	Flash           string              // used for flash messages
	IsAuthenticated bool                // used for authenticating user
	User            sqlc.GetUserByIDRow // used for account page
}

// newTemplateData returns a *templateData, which contains some fields having default values.
func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"), // The flash message is auto included the next time any page is rendered.
		IsAuthenticated: app.isAuthenticated(r),
	}
}

// initialTemplateCache parses all template files once when application is starting running,
// and storing those parsed template in an in-memory cache.
func initialTemplateCache() (map[string]*template.Template, error) {
	caches := make(map[string]*template.Template)

	// Get all relative file paths inside "pages" directory
	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		// Extract the file name from the full file path
		name := filepath.Base(page)

		// The template.FuncMap must be registered with the template set before you
		// call the ParseFiles() method. This means we have to use template.New() to
		// create an empty template set, use the Funcs() method to register the
		// template.FuncMap, and then parse the file as normal.
		ts, err := template.New(name).Funcs(functionTemplates).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		// Call ParseGlob() on ts to add any partials.
		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		// Call ParseFiles() on ts to add the page template.
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Add the template set to the map, using the name of the page
		// (like 'home.html') as the key
		caches[name] = ts
	}

	return caches, nil
}
