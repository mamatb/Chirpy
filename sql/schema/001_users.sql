-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    email text UNIQUE NOT NULL
);
CREATE TABLE chirps (
    id uuid PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    body text NOT NULL,
    user_id uuid REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE users;
DROP TABLE chirps;
