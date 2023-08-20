package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/chauvinhphuoc/snippetbox/internal/db/sqlc"
	"github.com/chauvinhphuoc/snippetbox/internal/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"strings"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	result, err := app.GetTenLatestSnippets(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
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

	result, err := app.GetSnippetNotExpired(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippet = result

	app.render(w, http.StatusOK, "view.html", data)
}

func (app *application) displayCreateSnippetForm(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
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
		data := app.newTemplateData(r)
		data.Form = form

		app.render(w, http.StatusUnprocessableEntity, "create-snippet.html", data)
		return
	}

	arg := sqlc.CreateSnippetParams{
		Title:    form.Title,
		Content:  form.Content,
		Duration: int32(form.Expires),
	}

	result, err := app.CreateSnippet(r.Context(), arg)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%v", result), http.StatusSeeOther)
}

func (app *application) displaySignupPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupFormResult{
		Name:      "",
		Email:     "",
		Password:  "",
		Validator: validator.Validator{},
	}

	app.render(w, http.StatusOK, "signup.html", data)
}

type userSignupFormResult struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) doCreateUser(w http.ResponseWriter, r *http.Request) {
	var form userSignupFormResult

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	// Validate the form contents using our helper functions.
	if !validator.IsNotBlank(form.Name) {
		form.AddFieldError("name", "This field cannot be blank")
	}

	if !validator.IsNotBlank(form.Email) {
		form.AddFieldError("email", "This field cannot be blank")
	}

	if !validator.IsMatchRegex(form.Email, validator.EmailRX) {
		form.AddFieldError("email", "This field must be a valid email address")
	}

	if !validator.IsNotBlank(form.Password) {
		form.AddFieldError("password", "This field cannot be blank")
	}

	if !validator.IsStringNotLessThanLimit(form.Password, 8) {
		form.AddFieldError("password", "This field must be at least 8 characters long")
	}

	// If there are any errors, redisplay the signup form along with a 422
	// status code.
	if !form.IsNoErrors() {
		data := app.newTemplateData(r)
		data.Form = form

		app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), 12)
	if err != nil {
		app.serverError(w, err)
		return
	}

	arg := sqlc.CreateUserParams{
		Name:           form.Name,
		Email:          form.Email,
		HashedPassword: string(hashedPassword),
	}

	err = app.CreateUser(r.Context(), arg)
	if err != nil {
		var postgreSQLError *pq.Error
		if errors.As(err, &postgreSQLError) {
			code := postgreSQLError.Code.Name()
			if code == "unique_violation" && strings.Contains(postgreSQLError.Message, "users_uc_email") {
				form.AddFieldError("email", "Email address is already in use")

				data := app.newTemplateData(r)
				data.Form = form
				app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
				return
			}
		}

		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	// Redirect user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) displayLoginPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginFormResult{
		Email:     "",
		Password:  "",
		Validator: validator.Validator{},
	}

	app.render(w, http.StatusOK, "login.html", data)
}

type userLoginFormResult struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) doLoginUser(w http.ResponseWriter, r *http.Request) {
	var form userLoginFormResult

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if !validator.IsNotBlank(form.Email) {
		form.AddFieldError("email", "This field cannot be blank")
	}

	if !validator.IsMatchRegex(form.Email, validator.EmailRX) {
		form.AddFieldError("email", "This field must be a valid email address")
	}

	if !validator.IsNotBlank(form.Password) {
		form.AddFieldError("password", "This field cannot be blank")
	}

	if !form.IsNoErrors() {
		data := app.newTemplateData(r)
		data.Form = form

		app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	// Check whether user with the email provided exists.
	user, err := app.GetUserByEmail(r.Context(), form.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			form.AddGenericFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, err)
		}

		return
	}

	// Check whether the hashed password and plain-text password that user provided match.
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(form.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			form.AddGenericFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, err)
		}

		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Add the ID of the current user to the session, so that they are now
	// 'logged in'.
	app.sessionManager.Put(r.Context(), "authenticatedUserID", user.ID)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) doLogoutUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Logout the user...")
}
