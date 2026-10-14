-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE users.name = $1;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT users.name
FROM users;

-- name: CreateFeeds :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeeds :many
SELECT 
feeds.name AS feed_name,
feeds.url,
users.name AS user_name
FROM feeds
INNER JOIN users
    ON feeds.user_id = users.id;

-- name: GetFeed :one
SELECT *
FROM feeds
WHERE feeds.url = $1;

-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
)
SELECT
inserted_feed_follow.*,
feeds.name AS feed_name,
users.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds
    ON inserted_feed_follow.feed_id = feeds.id
INNER JOIN users
    ON inserted_feed_follow.user_id = users.id;

-- name: GetFeedFollowsForUser :many
SELECT
feeds.name
FROM feed_follows
INNER JOIN feeds
    ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = $1;