-- +goose Up
SELECT 'up SQL query';
ALTER TABLE users ADD COLUMN birthday DATE;

-- +goose Down
SELECT 'down SQL query';
ALTER TABLE users DROP COLUMN birthday;
