-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS metrics (
    name TEXT NOT NULL,
    type SMALLINT,
    gauge DOUBLE PRECISION DEFAULT 0,
    counter BIGINT DEFAULT 0,
    PRIMARY KEY(name, type)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS metrics;
-- +goose StatementEnd
