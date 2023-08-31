-- name: CreateUser :exec
INSERT INTO users (name, email, hashed_password, created_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP);

-- name: GetUserByEmail :one
SELECT id, hashed_password
FROM users
WHERE email = $1;

-- name: IsUserExist :one
SELECT EXISTS(SELECT true FROM users WHERE id = $1);

-- name: GetUserByID :one
SELECT name, email, created_at FROM users WHERE id = $1;