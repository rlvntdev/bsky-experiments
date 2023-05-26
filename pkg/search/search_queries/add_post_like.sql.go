// Code generated by sqlc. DO NOT EDIT.
// source: add_post_like.sql

package search_queries

import (
	"context"
)

const addLikeToPost = `-- name: AddLikeToPost :exec
INSERT INTO post_likes (post_id, like_count)
VALUES ($1, 1)
ON CONFLICT (post_id)
DO UPDATE SET like_count = post_likes.like_count + 1
WHERE post_likes.post_id = $1
`

func (q *Queries) AddLikeToPost(ctx context.Context, postID string) error {
	_, err := q.exec(ctx, q.addLikeToPostStmt, addLikeToPost, postID)
	return err
}