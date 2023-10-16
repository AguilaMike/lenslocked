CREATE TABLE sessions (
    id UUID NOT NULL,
    user_id UUID NOT NULL,
    token_hash TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM now())::int,
    updated_at INTEGER,
    CONSTRAINT sessions_id_pk PRIMARY KEY (id),
    CONSTRAINT sessions_user_id_uq UNIQUE (user_id),
    CONSTRAINT rel_sessions_users_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE INDEX idx_session_token_hash ON sessions (token_hash);
