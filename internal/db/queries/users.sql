-- name: CreateUser :exec
INSERT INTO users (name, email, hashed_password, created_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP);

-- name: GetUserByEmail :one
SELECT id, hashed_password FROM users WHERE email = $1;