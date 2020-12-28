CREATE TABLE users (
	name          TEXT PRIMARY KEY CHECK(name <> ''),
	password_hash TEXT NOT NULL CHECK(password_hash <> ''),
	avatar        BLOB
);

CREATE TABLE rooms (
	name TEXT PRIMARY KEY CHECK(name <> '')
);

CREATE TABLE messages (
	id         INTEGER PRIMARY KEY,
	content    TEXT NOT NULL CHECK(content <> ''),
	created_at TEXT NOT NULL CHECK(created_at <> ''), -- rfc3339: 2019-10-25T07:55:50Z
	room       TEXT NOT NULL REFERENCES rooms(name),
	author     TEXT NOT NULL REFERENCES users(name)
);

CREATE TABLE schema (
	version TEXT PRIMARY KEY CHECK(version <> '')
);

INSERT INTO schema VALUES ('v0.0.1');
