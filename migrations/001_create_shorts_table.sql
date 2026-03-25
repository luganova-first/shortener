-- +goose Up
CREATE TABLE IF NOT EXISTS shorts  (
    id bigint GENERATED ALWAYS AS IDENTITY,
    shorted VARCHAR(8) NOT NULL UNIQUE DEFAULT '',
    full_url VARCHAR(250) NOT NULL DEFAULT '',
    is_deleted BOOL NOT NULL DEFAULT FALSE,
    PRIMARY KEY(id)
);

-- +goose Down
DROP TABLE IF EXISTS shorts;