CREATE TABLE IF NOT EXISTS users (
	name          TEXT PRIMARY KEY,
	password_hash TEXT NOT NULL,
	avatar        BLOB
);

CREATE TABLE IF NOT EXISTS rooms (
	name TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS messages (
	id         INTEGER PRIMARY KEY,
	content    TEXT NOT NULL,
	created_at TEXT NOT NULL, -- rfc3339: 2019-10-25T07:55:50Z
	room       TEXT NOT NULL REFERENCES rooms(name),
	author     TEXT NOT NULL REFERENCES users(name)
);

CREATE TABLE IF NOT EXISTS schema (
	version TEXT PRIMARY KEY
);

-- Only insert version into empty schema table.
INSERT INTO schema SELECT 'v0.0.1' WHERE NOT EXISTS (SELECT * FROM schema);
