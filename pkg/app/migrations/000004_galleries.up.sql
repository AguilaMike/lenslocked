CREATE TABLE galleries (
  id UUID NOT NULL,
  user_id UUID NOT NULL,
  title TEXT,
  created_at INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM now())::int,
  updated_at INTEGER,
  CONSTRAINT galleries_id_pk PRIMARY KEY (id),
  CONSTRAINT galleries_user_id_title_uq UNIQUE (user_id, title)
);

CREATE INDEX idx_galleries_id ON galleries (id);
CREATE INDEX idx_galleries_user_id ON galleries (user_id);
CREATE INDEX idx_galleries_user_id_title ON galleries (user_id, title);
