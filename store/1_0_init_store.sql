-- +migrate Up
CREATE TABLE users(
  id		 SERIAL,
  name varchar NOT NULL UNIQUE,
  access_token varchar NOT NULL,
  status int4 NOT NULL DEFAULT 1,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  usage int8 NOT NULL DEFAULT 0,
  usage_limit int8 NOT NULL DEFAULT 0,
  constraint users_pk primary key(id)
);

-- test API user
INSERT INTO users(name, access_token) VALUES('test', 'test');

CREATE TABLE queue(
    id		 SERIAL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    processed_at timestamp,
    params jsonb,
    user_id int4 NOT NULL,
    action varchar NOT NULL,
    error jsonb,
    result jsonb,
    next_try_at int8,
    try_count int8,
    constraint queue_pk primary key(id)
);

CREATE TABLE entries(
    id		 SERIAL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    entryhash VARCHAR(64) UNIQUE NOT NULL,
    entrydata jsonb,
    constraint entry_pk primary key(id)
);

CREATE TABLE chains(
    chainid VARCHAR(64) UNIQUE NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    content TEXT NOT NULL,
    extids jsonb NOT NULL
);

CREATE TABLE users_chains(
    chainid VARCHAR(64) NOT NULL,
    user_id int8 NOT NULL,
    constraint users_chains_pk primary key (chainid, user_id)
);

-- +migrate Down
DROP TABLE users;
DROP TABLE queue;
DROP TABLE entries;
DROP TABLE chains;
DROP TABLE users_chains;