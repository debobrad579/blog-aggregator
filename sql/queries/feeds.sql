-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
    VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: GetFeed :one
SELECT
    *
FROM
    feeds
WHERE
    url = $1;

-- name: GetFeeds :many
SELECT
    feeds.*,
    users.name AS user_name
FROM
    feeds
    JOIN users ON users.id = feeds.user_id
ORDER BY
    feeds.name;

-- name: MarkFeedFetched :exec
UPDATE
    feeds
SET
    updated_at = $1,
    last_fetched_at = $1
WHERE
    id = $2;

-- name: GetNextFeedToFetch :one
SELECT
    *
FROM
    feeds
ORDER BY
    last_fetched_at NULLS FIRST
LIMIT 1;

