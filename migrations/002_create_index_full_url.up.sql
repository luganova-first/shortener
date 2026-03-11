-- Создание уникального индекса для поля full_url
CREATE UNIQUE INDEX idx_shorts_full_url_unique ON shorts (full_url);
