-- +migrate Up
ALTER TABLE users DROP CONSTRAINT users_name_key;

-- +migrate Down
ALTER TABLE users ADD CONSTRAINT users_name_key UNIQUE (name);