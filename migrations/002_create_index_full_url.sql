-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS idx_shorts_full_url_unique ON shorts (full_url);
