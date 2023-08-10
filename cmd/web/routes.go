package main

import "net/http"

// routes() returns a server multiplexer aka a router of interface type http.Handler containing all application routes.
// By implementing like this, first, we only can register handlers in this routes function and assign that handlers to the server in the main function.
// Second, we treat the mux like other handlers, so we can easily implement middlewares in the future.
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippet/view", app.viewSnippet)
	mux.HandleFunc("/snippet/create", app.createSnippet)
	return app.logRequest(mux)
}
