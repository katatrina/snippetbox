This application lets people paste and share snippets of text - a bit like [Pastebin](https://pastebin.pl/) or
GitHub's [Gist](https://gist.github.com/).

## Prerequisites

- Go >= 1.20
- PostgresSQL 15

## Database configurations

## Application setups

## Workflow

## Available routes

| Method | Pattern           | Handler                  | Action                                         |
|--------|-------------------|--------------------------|------------------------------------------------|
| GET    | /                 | home                     | Display the home page                          |
| GET    | /snippet/view/:id | viewSnippet              | Display a specific snippet                     |
| GET    | /snippet/create   | displayCreateSnippetForm | Display a HTML form for creating a new snippet |
| POST   | /snippet/create   | doCreateSnippet          | Create a new snippet                           |
| GET    | /user/signup      | displaySignupPage        | Display a HTMl form for signing up a new user  |
| POST   | /user/signup      | doSignupUser             | Create a new user                              |
| GET    | /user/login       | displayLoginPage         | Display a HTMl form for logging in a user      |
| POST   | /user/login       | doLoginUser              | Authenticate and login the user                |
| POST   | /user/logout      | doLogoutUser             | Logout the user                                |
| GET    | /static/*filepath | http.FileServer          | Serve a specific static file                   |