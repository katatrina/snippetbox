// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: users.sql

package sqlc

import (
	"context"
	"time"
)

const createUser = `-- name: CreateUser :exec
INSERT INTO users (name, email, hashed_password, created_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
`

type CreateUserParams struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	HashedPassword string `json:"hashed_password"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) error {
	_, err := q.db.ExecContext(ctx, createUser, arg.Name, arg.Email, arg.HashedPassword)
	return err
}

const getPasswordByID = `-- name: GetPasswordByID :one
SELECT hashed_password
FROM users
WHERE id = $1
`

func (q *Queries) GetPasswordByID(ctx context.Context, id int32) (string, error) {
	row := q.db.QueryRowContext(ctx, getPasswordByID, id)
	var hashed_password string
	err := row.Scan(&hashed_password)
	return hashed_password, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, hashed_password
FROM users
WHERE email = $1
`

type GetUserByEmailRow struct {
	ID             int32  `json:"id"`
	HashedPassword string `json:"hashed_password"`
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i GetUserByEmailRow
	err := row.Scan(&i.ID, &i.HashedPassword)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT name, email, created_at
FROM users
WHERE id = $1
`

type GetUserByIDRow struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (q *Queries) GetUserByID(ctx context.Context, id int32) (GetUserByIDRow, error) {
	row := q.db.QueryRowContext(ctx, getUserByID, id)
	var i GetUserByIDRow
	err := row.Scan(&i.Name, &i.Email, &i.CreatedAt)
	return i, err
}

const isUserExist = `-- name: IsUserExist :one
SELECT EXISTS(SELECT true FROM users WHERE id = $1)
`

func (q *Queries) IsUserExist(ctx context.Context, id int32) (bool, error) {
	row := q.db.QueryRowContext(ctx, isUserExist, id)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const updateUserPassword = `-- name: UpdateUserPassword :exec
UPDATE users
SET hashed_password = $1
WHERE id = $2
`

type UpdateUserPasswordParams struct {
	HashedPassword string `json:"hashed_password"`
	ID             int32  `json:"id"`
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error {
	_, err := q.db.ExecContext(ctx, updateUserPassword, arg.HashedPassword, arg.ID)
	return err
}
