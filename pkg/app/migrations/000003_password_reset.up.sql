CREATE TABLE password_resets (
  id UUID NOT NULL,
  user_id UUID NOT NULL,
  token_hash TEXT NOT NULL,
  expires_at INTEGER NOT NULL,
  created_at INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM now())::int,
  updated_at INTEGER,
  CONSTRAINT password_resets_id_pk PRIMARY KEY (id),
  CONSTRAINT password_resets_user_id_uq UNIQUE (user_id),
  CONSTRAINT password_resets_token_hash_uq UNIQUE (token_hash),
  CONSTRAINT rel_password_resets_users_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_password_resets_id ON password_resets (id);
CREATE INDEX idx_password_resets_user_id ON password_resets (user_id);
CREATE INDEX idx_password_resets_token_hash ON password_resets (token_hash);
