CREATE TABLE users
(
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(255) NOT NULL,
    hashed_password CHAR(60)     NOT NULL,
    created_at      timestamptz  NOT NULL
);

ALTER TABLE users
    ADD CONSTRAINT users_uc_email UNIQUE (email);