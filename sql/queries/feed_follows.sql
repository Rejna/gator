-- name: CreateFeedFollow :one
WITH inserted as (INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *)
SELECT i.*, u.name, f.name
FROM inserted i
JOIN users u ON i.user_id = u.id
JOIN feeds f ON i.feed_id = f.id;

-- name: GetFeedFollowsForUser :many
SELECT u.name AS username, f.name AS feed_name, f.url AS feed_url
FROM feed_follows ff
JOIN users u ON ff.user_id = u.id
JOIN feeds f ON ff.feed_id = f.id
WHERE u.name = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows ff
USING feeds f
WHERE ff.user_id = $1 AND f.id = ff.feed_id AND f.url = $2;