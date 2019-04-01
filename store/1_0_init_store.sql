-- +migrate Up
CREATE TABLE users(
  id		 SERIAL,
  name VARCHAR(128) UNIQUE NOT NULL,
  access_token VARCHAR(128) UNIQUE NOT NULL,
  status INT4 NOT NULL DEFAULT 1,
  usage INT8 NOT NULL DEFAULT 0,
  usage_limit INT8 NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ,
  CONSTRAINT users_id_key PRIMARY KEY(id)
);

-- test API user
INSERT INTO users(name, access_token) VALUES('test', 'test');

CREATE TABLE queue(
    id		 SERIAL,
    user_id INT4 NOT NULL,
    action VARCHAR(32) NOT NULL,
    params BYTEA,
    error TEXT,
    result VARCHAR(64),
    processed_at TIMESTAMPTZ,
    next_try_at TIMESTAMPTZ,
    try_count INT4,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT queue_id_key PRIMARY KEY(id),
    CONSTRAINT queue_user_id_fkey FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE chains(
    chain_id VARCHAR(64) UNIQUE NOT NULL,
    -- content TEXT,
    ext_ids _TEXT,
    status VARCHAR(32),
    synced BOOLEAN NOT NULL DEFAULT FALSE,
    earliest_entry_block VARCHAR(64),
    latest_entry_block VARCHAR(64),
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chains_chain_id_key PRIMARY KEY (chain_id)
);

CREATE TABLE e_blocks(
    key_mr VARCHAR(64) UNIQUE NOT NULL,
    block_sequence_number INT8,
    chain_id VARCHAR(64),
    prev_key_mr VARCHAR(64),
    timestamp INT8,
    db_height INT8,
    CONSTRAINT e_blocks_key_mr_key PRIMARY KEY(key_mr)
);

CREATE TABLE entries(
    entry_hash VARCHAR(64) UNIQUE NOT NULL,
    chain_id VARCHAR(64),
    content TEXT,
    ext_ids _TEXT,
    status VARCHAR(32),
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT entries_entry_hash_key PRIMARY KEY(entry_hash),
    CONSTRAINT entries_chain_id_fkey FOREIGN KEY(chain_id) REFERENCES chains(chain_id)
);

CREATE TABLE entries_e_blocks(
    entry_entry_hash VARCHAR(64) NOT NULL,
    e_block_key_mr VARCHAR(64) NOT NULL,
    CONSTRAINT entries_e_blocks_pk PRIMARY KEY (entry_entry_hash, e_block_key_mr),
    CONSTRAINT entries_e_blocks_entry_hash_fkey FOREIGN KEY (entry_entry_hash) REFERENCES entries(entry_hash),
    CONSTRAINT entries_e_blocks_key_mr_fkey FOREIGN KEY (e_block_key_mr) REFERENCES e_blocks(key_mr)
);

CREATE TABLE users_chains(
    chain_chain_id VARCHAR(64) NOT NULL,
    user_id int8 NOT NULL,
    CONSTRAINT users_chains_pk PRIMARY KEY (chain_chain_id, user_id),
    CONSTRAINT users_chains_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT users_chains_chain_chain_id_fkey FOREIGN KEY (chain_chain_id) REFERENCES chains(chain_id)
);

-- +migrate Down
DROP TABLE users;
DROP TABLE queue;
DROP TABLE entries;
DROP TABLE chains;
DROP TABLE users_chains;