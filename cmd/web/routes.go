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

	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/snippet/view/:id", app.viewSnippet)
	router.HandlerFunc(http.MethodGet, "/snippet/create", app.displayCreateSnippetForm)
	router.HandlerFunc(http.MethodPost, "/snippet/create", app.doCreateSnippet)

	return app.logRequest(router)
}
