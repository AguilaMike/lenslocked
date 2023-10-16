CREATE TABLE users (
	id UUID NOT NULL,
	email TEXT NOT NULL,
	email_normalized TEXT NOT NULL,
	password_hash TEXT NOT NULL,
	is_admin BOOL NOT NULL DEFAULT FALSE,
	details JSONB NOT NULL DEFAULT '{}',
	created_at INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM now())::int,
	updated_at INTEGER,
	CONSTRAINT users_id_pk PRIMARY KEY (id),
	CONSTRAINT users_email_uq UNIQUE (email)
);
CREATE INDEX idx_users_id ON users (id);
CREATE INDEX idx_users_email ON users (email);
