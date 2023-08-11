This application lets people paste and share snippets of text - a bit like [Pastebin](https://pastebin.pl/) or GitHub's [Gist](https://gist.github.com/).

## Prerequisites

- Go >= 1.20
- PostgreSQL 15

## Available routes

| Method | Pattern           | Handler           | Action                                         |
|--------|-------------------|-------------------|------------------------------------------------|
| GET    | /                 | home              | Display the home page                          |
| GET    | /snippet/view/:id | viewSnippet       | Display a specific snippet                     |
| GET    | /snippet/create   | createSnippet     | Display a HTML form for creating a new snippet |
| POST   | /snippet/create   | createSnippetPost | Create a new snippet                           |
| GET    | /static/*filepath | http.FileServer   | Serve a specific static file                   |  