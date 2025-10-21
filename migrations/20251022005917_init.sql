-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS practicum.metrics (
    name TEXT NOT NULL,
    type SMALLINT,
    gauge DOUBLE PRECISION DEFAULT 0,
    counter BIGINT DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS practicum.metrics;
-- +goose StatementEnd
