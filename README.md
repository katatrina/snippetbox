This application lets people paste and share snippets of text - a bit like [Pastebin](https://pastebin.pl/) or
GitHub's [Gist](https://gist.github.com/).
This project
uses [Session-Cookie Authentication](https://blog.bytebytego.com/i/112781858/session-cookie-authentication).

## Prerequisites

- Go >= 1.20
- PostgreSQL 15

## Setups

- Type `go mod tidy` to install all project dependencies.
- Configure database source name and schema in files **cmd/web/main.go** and **internal/db/migrations**.
- Type `go run ./cmd/web` to start application.

* Optional: You can find shorter commands in Makefile.

## Available routes

| Method | Pattern           | Handler                  | Action                                         |
|--------|-------------------|--------------------------|------------------------------------------------|
| GET    | /                 | home                     | Display the home page                          |
| GET    | /snippet/view/:id | viewSnippet              | Display a specific snippet                     |
| GET    | /snippet/create   | displayCreateSnippetForm | Display a HTML form for creating a new snippet |
| POST   | /snippet/create   | doCreateSnippet          | Create a new snippet                           |
| GET    | /user/signup      | displaySignupPage        | Display a HTML form for signing up a new user  |
| POST   | /user/signup      | doSignupUser             | Create a new user                              |
| GET    | /user/login       | displayLoginPage         | Display a HTML form for logging in a user      |
| POST   | /user/login       | doLoginUser              | Authenticate and login the user                |
| POST   | /user/logout      | doLogoutUser             | Logout the user                                |
| GET    | /static/*filepath | http.FileServer          | Serve a specific static file                   |
| GET    | /account/view     | viewAccount              | View account's information for each user       |
| GET    | /about            | about                    | Display the about page                         |