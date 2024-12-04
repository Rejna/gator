-- +goose Up
CREATE TABLE posts (
 id UUID PRIMARY KEY,
 created_at TIMESTAMP NOT NULL,
 updated_at TIMESTAMP NOT NULL,
 title VARCHAR(255) NOT NULL,
 url VARCHAR(255) NOT NULL UNIQUE,
 description TEXT NULL,
 published_at TIMESTAMP NULL,
 feed_id UUID NOT NULL REFERENCES feeds on DELETE CASCADE,
 FOREIGN KEY(feed_id) REFERENCES feeds (id)
);

-- +goose Down
DROP TABLE posts;