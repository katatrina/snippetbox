package main

import (
	"database/sql"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/chauvinhphuoc/snippetbox/internal/db/sqlc"
	"github.com/go-playground/form/v4"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var (
	infoLog        = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog       = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	dataSourceName = "postgres://postgres:12345@localhost:5432/snippetbox?sslmode=disable"
)

// Everything inside an application is called a dependency,
// it sticks to the application for doing tasks.
type application struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	db       *sql.DB // In case of executing a transaction.
	*sqlc.Queries
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder // A Decoder instance is used to map HTML field values into struct fields.
	sessionManager *scs.SessionManager
}

func main() {
	db, err := openDB()
	if err != nil {
		errorLog.Fatal(err)
	}
	infoLog.Print("Connected to database")

	q := sqlc.New(db)

	templateCache, err := initialTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	// Configure the session manager to use our PostgreSQL database as the session store (in the table "sessions").
	sessionManager.Store = postgresstore.New(db)
	// Sessions automatically expire 12 hours after first being created.
	sessionManager.Lifetime = 12 * time.Hour

	app := &application{
		infoLog:        infoLog,
		errorLog:       errorLog,
		db:             db,
		Queries:        q,
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	server := &http.Server{
		Addr:    "127.0.0.1:4000",
		Handler: app.routes(),
	}

	infoLog.Print("Starting server on http://localhost:4000")
	err = server.ListenAndServe()
	errorLog.Print(err)
}

func openDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	if pingErr := db.Ping(); pingErr != nil {
		return nil, err
	}

	return db, nil
}
