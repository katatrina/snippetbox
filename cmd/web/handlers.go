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

// GET /
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.GetTenLatestSnippets(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippets = snippets

	app.render(w, http.StatusOK, "home.html", data)
}

// GET /snippet/view/:id
func (app *application) viewSnippet(w http.ResponseWriter, r *http.Request) {
	// params are parameters from URL path, not query parameters
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	snippet, err := app.GetSnippetNotExpired(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippet = snippet

	app.render(w, http.StatusOK, "view.html", data)
}

// GET /snippet/create
func (app *application) displayCreateSnippetForm(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = createSnippetFormResult{
		// Other fields get zero-value.
		Expires: 365, // The value "One year" of radio button "Delete in" is chosen by default.
	}

	app.render(w, http.StatusOK, "create-snippet.html", data)
}

// createSnippetFormResult represents the form data and validation errors
// for the form fields.
type createSnippetFormResult struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

// POST /snippet/create
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

	snippet, err := app.CreateSnippet(r.Context(), arg)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Create snippet successfully.")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%v", snippet), http.StatusSeeOther)
}

// GET /user/signup
func (app *application) displaySignupPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupFormResult{}

	app.render(w, http.StatusOK, "signup.html", data)
}

type userSignupFormResult struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

// POST /user/signup
func (app *application) doSignupUser(w http.ResponseWriter, r *http.Request) {
	var form userSignupFormResult

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
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

	// Else, try to insert user's information to database.

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

// GET /user/login
func (app *application) displayLoginPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginFormResult{}

	app.render(w, http.StatusOK, "login.html", data)
}

type userLoginFormResult struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

// POST /user/login
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
			form.AddGenericError("Email or password is incorrect")

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
			form.AddGenericError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, err)
		}

		return
	}

	// Until here, the user is authenticated successfully.

	// Use the RenewToken() method on the current session to change the session
	// ID. It's good practice to generate a new session ID when the
	// authentication state or privilege levels changes for the user (e.g. login
	// and logout operations).
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Add the ID of the current user to the session, so that they are now
	// 'logged in'.
	app.sessionManager.Put(r.Context(), "authenticatedUserID", int(user.ID))

	path := app.sessionManager.PopString(r.Context(), "redirectPathAfterLogin")
	if path != "" {
		http.Redirect(w, r, path, http.StatusSeeOther)
		return
	}

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

// POST /user/logout
func (app *application) doLogoutUser(w http.ResponseWriter, r *http.Request) {
	// Use the RenewToken() method on the current session to change the session
	// ID again.
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Remove the authenticatedUserID from the session data so that user is
	// 'logged out'.
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	// Add a flash message to the session to confirm to the user that they've been
	// logged out.
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// GET /about
func (app *application) about(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, http.StatusOK, "about.html", data)
}

// GET /account/view
func (app *application) viewAccount(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	user, err := app.GetUserByID(r.Context(), int32(userID))
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.User = user

	app.render(w, http.StatusOK, "account.html", data)
}
