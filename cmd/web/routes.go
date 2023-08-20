package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// routes() returns a server multiplexer aka a router of interface type http.Handler containing all application routes.
// By implementing like this, first, we only can register handlers in this routes method and assign that handlers to the server in the main function.
// Second, we treat the router like other handlers, so we can easily implement middlewares in the future.
func (app *application) routes() http.Handler {
	router := httprouter.New()

	fileServer := http.FileServer(http.Dir("./ui/static"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static/", fileServer))

	//router.HandlerFunc(http.MethodGet, "/", app.home)
	//router.HandlerFunc(http.MethodGet, "/snippet/view/:id", app.viewSnippet)
	//router.HandlerFunc(http.MethodGet, "/snippet/create", app.displayCreateSnippetForm)
	//router.HandlerFunc(http.MethodPost, "/snippet/create", app.doCreateSnippet)

	router.Handler(http.MethodGet, "/", app.sessionManager.LoadAndSave(http.HandlerFunc(app.home)))
	router.Handler(http.MethodGet, "/snippet/view/:id", app.sessionManager.LoadAndSave(http.HandlerFunc(app.viewSnippet)))
	router.Handler(http.MethodGet, "/snippet/create", app.sessionManager.LoadAndSave(http.HandlerFunc(app.displayCreateSnippetForm)))
	router.Handler(http.MethodPost, "/snippet/create", app.sessionManager.LoadAndSave(http.HandlerFunc(app.doCreateSnippet)))

	router.Handler(http.MethodGet, "/user/signup", app.sessionManager.LoadAndSave(http.HandlerFunc(app.displaySignupPage)))
	router.Handler(http.MethodPost, "/user/signup", app.sessionManager.LoadAndSave(http.HandlerFunc(app.doCreateUser)))
	router.Handler(http.MethodGet, "/user/login", app.sessionManager.LoadAndSave(http.HandlerFunc(app.displayLoginPage)))
	router.Handler(http.MethodPost, "/user/login", app.sessionManager.LoadAndSave(http.HandlerFunc(app.doLoginUser)))
	router.Handler(http.MethodPost, "/user/logout", app.sessionManager.LoadAndSave(http.HandlerFunc(app.doLogoutUser)))

	return app.logRequest(router)
}
