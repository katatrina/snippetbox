package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

// routes() returns a server multiplexer aka a router of interface type http.Handler containing all application routes.
// By implementing like this, first, we only can register handlers in this routes method and assign that handlers to the server in the main function.
// Second, we treat the router like other handlers, so we can easily implement middlewares in the future.
func (app *application) routes() http.Handler {
	router := httprouter.New()

	fileServer := http.FileServer(http.Dir("./ui/static"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static/", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave, app.authenticate)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.viewSnippet))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.displaySignupPage))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.doSignupUser))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.displayLoginPage))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.doLoginUser))
	router.Handler(http.MethodGet, "/about", dynamic.ThenFunc(app.about))

	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.displayCreateSnippetPage))
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.doCreateSnippet))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.doLogoutUser))
	router.Handler(http.MethodGet, "/account/view", protected.ThenFunc(app.viewAccount))
	router.Handler(http.MethodGet, "/account/change-password", protected.ThenFunc(app.displayChangeUserPasswordPage))
	router.Handler(http.MethodPost, "/account/change-password", protected.ThenFunc(app.doUpdateUserPassword))

	standard := alice.New(app.logRequest)
	return standard.Then(router)
}
