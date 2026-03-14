-- +goose Up
CREATE TABLE IF NOT EXISTS shorts  (
    id bigint GENERATED ALWAYS AS IDENTITY,
    shorted VARCHAR(8) NOT NULL DEFAULT '',
    full_url VARCHAR(250) NOT NULL DEFAULT '',
    PRIMARY KEY(id)
);

-- +goose Down
DROP TABLE IF EXISTS shorts;