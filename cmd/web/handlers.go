package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/chauvinhphuoc/snippetbox/internal/db/sqlc"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	result, err := app.GetTenLatestSnippets(context.Background())
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := app.newTemplateData()
	data.Snippets = result

	filenames := []string{
		"./ui/html/pages/home.html",
		"./ui/html/base.html",
		"./ui/html/partials/navbar.html",
	}

	// The template.FuncMap must be registered with the template set before you
	// call the ParseFiles() method.
	ts := template.New(r.URL.Path).Funcs(map[string]any{
		"humanDate": humanDate,
	})
	ts, err = ts.ParseFiles(filenames...)

	if err != nil {
		app.serverError(w, err)
		return
	}

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) viewSnippet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	result, err := app.GetSnippetNotExpired(context.Background(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		app.errorLog.Print(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	files := []string{
		"./ui/html/base.html",
		"./ui/html/pages/view.html",
		"./ui/html/partials/navbar.html",
	}

	// The template.FuncMap must be registered with the template set before you
	// call the ParseFiles() method.
	ts := template.New(r.URL.Path).Funcs(map[string]any{
		"humanDate": humanDate,
	})
	ts, err = ts.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData()
	data.Snippet = result

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	arg := sqlc.CreateSnippetParams{
		Title:   "Amazing Spider Man",
		Content: "Today is a beautiful day",
		Expires: time.Now().AddDate(0, 0, 7),
	}

	result, err := app.CreateSnippet(context.Background(), arg)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet/view?id=%v", result), http.StatusSeeOther)
}
