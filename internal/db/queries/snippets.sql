-- name: CreateSnippet :one
INSERT INTO snippets (title, content, created_at, expires)
VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + MAKE_INTERVAL(DAYS => sqlc.arg(duration)::int)) RETURNING id;

-- name: GetSnippetNotExpired :one
SELECT *
FROM snippets
WHERE expires > CURRENT_TIMESTAMP
  AND id = $1;

-- name: GetTenLatestSnippets :many
SELECT *
FROM snippets
WHERE expires > CURRENT_TIMESTAMP
ORDER BY id DESC LIMIT 10;
