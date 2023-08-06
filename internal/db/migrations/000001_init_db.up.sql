CREATE TABLE snippets
(
    id         serial PRIMARY KEY,
    title      varchar(100) NOT NULL,
    content    text         NOT NULL,
    created_at timestamptz  NOT NULL DEFAULT NOW(),
    expires    timestamptz  NOT NULL
);

CREATE INDEX ON snippets (created_at);