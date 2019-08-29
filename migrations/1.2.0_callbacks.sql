-- +migrate Up
CREATE TABLE callbacks(
    id		 SERIAL,
    user_id INT4 NOT NULL,
    entry_hash VARCHAR(64) NOT NULL,
    url VARCHAR,
    result INT4,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT callbacks_id_key PRIMARY KEY(id),
    CONSTRAINT callbacks_user_id_fkey FOREIGN KEY(user_id) REFERENCES users(id),
    CONSTRAINT callbacks_entry_hash_fkey FOREIGN KEY(entry_hash) REFERENCES entries(entry_hash)
);

-- +migrate Down
DROP TABLE callbacks;