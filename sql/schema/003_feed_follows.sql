-- +goose Up
CREATE TABLE feed_follows (
 id UUID PRIMARY KEY,
 created_at TIMESTAMP NOT NULL,
 updated_at TIMESTAMP NOT NULL,
 user_id UUID NOT NULL REFERENCES users on DELETE CASCADE,
 feed_id UUID NOT NULL REFERENCES feeds on DELETE CASCADE,
 UNIQUE(user_id, feed_id),
 FOREIGN KEY(user_id) REFERENCES users (id),
 FOREIGN KEY(feed_id) REFERENCES feeds (id)
);

-- +goose Down
DROP TABLE feed_follows;