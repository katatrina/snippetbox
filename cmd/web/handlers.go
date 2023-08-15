package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/chauvinhphuoc/snippetbox/internal/db/sqlc"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	result, err := app.GetTenLatestSnippets(context.Background())
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := newTemplateData()
	data.Snippets = result

	app.render(w, http.StatusOK, "home.html", data)
}

func (app *application) viewSnippet(w http.ResponseWriter, r *http.Request) {
	// params are parameters from URL path, not query parameters
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	result, err := app.GetSnippetNotExpired(context.Background(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, err)
		return
	}

	files := []string{
		"ui/html/base.html",
		"ui/html/pages/view.html",
		"ui/html/partials/navbar.html",
	}

	ts := template.New(r.URL.Path).Funcs(map[string]any{
		"humanDate": humanDate,
	})
	ts, err = ts.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := newTemplateData()
	data.Snippet = result

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) displayCreateSnippetForm(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.html",
		"./ui/html/pages/create-snippet.html",
		"./ui/html/partials/navbar.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		app.serverError(w, err)
	}
}

// createSnippetForm represents the form data and validation errors
// for the form fields.
type createSnippetForm struct {
	Title       string
	Content     string
	Expires     int32
	FieldErrors map[string]string
}

func (app *application) createSnippetPost(w http.ResponseWriter, r *http.Request) {
	// r.ParseForm() adds any data in POST request bodies to the r.PostForm map.
	err := r.ParseForm()
	if err != nil {
		// I think we need logging here because err may be due to either a server error or client error.
		app.errorLog.Print(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := createSnippetForm{
		Title:       r.PostForm.Get("title"),
		Content:     r.PostForm.Get("content"),
		Expires:     int32(expires),
		FieldErrors: map[string]string{},
	}

	if strings.TrimSpace(form.Title) == "" {
		form.FieldErrors["title"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(form.Title) > 100 {
		form.FieldErrors["title"] = "This field cannot be more than 100 characters long"
	}

	if strings.TrimSpace(form.Content) == "" {
		form.FieldErrors["content"] = "This field cannot be blank"
	}

	if expires != 1 && expires != 7 && expires != 365 {
		form.FieldErrors["expires"] = "This field must equal 1, 7 or 365"
	}

	// If there are any validation errors, re-display the create-snippet.html with error notifications.
	if len(form.FieldErrors) > 0 {
		data := newTemplateData()
		data.Form = form

		files := []string{
			"./ui/html/base.html",
			"./ui/html/pages/create-snippet.html",
			"./ui/html/partials/navbar.html",
		}

		ts, err := template.ParseFiles(files...)
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusUnprocessableEntity)
		err = ts.ExecuteTemplate(w, "base", data)
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		return
	}

	arg := sqlc.CreateSnippetParams{
		Title:    form.Title,
		Content:  form.Content,
		Duration: int32(expires),
	}

	result, err := app.CreateSnippet(context.Background(), arg)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%v", result), http.StatusSeeOther)
}
