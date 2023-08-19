package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/chauvinhphuoc/snippetbox/internal/db/sqlc"
	"github.com/chauvinhphuoc/snippetbox/internal/validator"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"net/http"
	"strconv"
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

	flash := app.sessionManager.PopString(r.Context(), "flash")

	data := newTemplateData()
	data.Snippet = result
	data.Flash = flash

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) displayCreateSnippetForm(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData()
	data.Form = createSnippetFormResult{
		Title:   "",
		Content: "",
		Expires: 365, // The value "One year" of radio button "Delete in" is chosen by default.
		//Validator: validator.Validator{FieldErrors: nil}, <- this is zero-value
	}

	app.render(w, http.StatusOK, "create-snippet.html", data)
}

// createSnippetForm represents the form data and validation errors
// for the form fields.
type createSnippetFormResult struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

func (app *application) doCreateSnippet(w http.ResponseWriter, r *http.Request) {
	var form createSnippetFormResult

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// validate title
	if !validator.IsNotBlank(form.Title) {
		form.AddFieldError("title", "This field cannot be blank")
	}
	if !validator.IsStringNotExceedLimit(form.Title, 100) {
		form.AddFieldError("title", "This field cannot be more than 100 characters")
	}

	// validate content
	if !validator.IsNotBlank(form.Content) {
		form.AddFieldError("content", "This field cannot be blank")
	}

	// validate expires
	if !validator.IsIntInList(form.Expires, 1, 7, 365) {
		form.AddFieldError("expires", "This field must equal 1, 7 or 365")
	}

	// If there are any validation errors, re-display the create-snippet.html with error notifications.
	// The URL path still does not change.
	if !form.IsNoErrors() {
		data := newTemplateData()
		data.Form = form

		app.render(w, http.StatusUnprocessableEntity, "create-snippet.html", data)
		return
	}

	arg := sqlc.CreateSnippetParams{
		Title:    form.Title,
		Content:  form.Content,
		Duration: int32(form.Expires),
	}

	result, err := app.CreateSnippet(context.Background(), arg)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%v", result), http.StatusSeeOther)
}
