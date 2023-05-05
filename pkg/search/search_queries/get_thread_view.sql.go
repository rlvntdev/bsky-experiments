// Code generated by sqlc. DO NOT EDIT.
// source: get_thread_view.sql

package search_queries

import (
	"context"
	"database/sql"
	"time"
)

const getThreadView = `-- name: GetThreadView :many
WITH RECURSIVE post_tree AS (
    -- Base case: select root post
    SELECT posts.id,
           text,
           parent_post_id,
           root_post_id,
           author_did,
           a2.handle,
           created_at,
           has_embedded_media,
           0 AS depth
    FROM posts
             LEFT JOIN authors a2 on a2.did = posts.author_did
    WHERE posts.id = $1
      AND posts.author_did = $2

    UNION ALL

    -- Recursive case: select child posts
    SELECT p2.id,
           p2.text,
           p2.parent_post_id,
           p2.root_post_id,
           p2.author_did,
           a.handle,
           p2.created_at,
           p2.has_embedded_media,
           pt.depth + 1 AS depth
    FROM posts p2
             JOIN
         post_tree pt ON p2.parent_post_id = pt.id AND p2.parent_relationship = 'r'
             LEFT JOIN authors a on p2.author_did = a.did)

SELECT id,
       text,
       parent_post_id,
       root_post_id,
       author_did,
       handle,
       created_at,
       has_embedded_media,
       depth
FROM post_tree
ORDER BY depth
`

type GetThreadViewParams struct {
	ID        string `json:"id"`
	AuthorDid string `json:"author_did"`
}

type GetThreadViewRow struct {
	ID               string         `json:"id"`
	Text             string         `json:"text"`
	ParentPostID     sql.NullString `json:"parent_post_id"`
	RootPostID       sql.NullString `json:"root_post_id"`
	AuthorDid        string         `json:"author_did"`
	Handle           sql.NullString `json:"handle"`
	CreatedAt        time.Time      `json:"created_at"`
	HasEmbeddedMedia bool           `json:"has_embedded_media"`
	Depth            interface{}    `json:"depth"`
}

func (q *Queries) GetThreadView(ctx context.Context, arg GetThreadViewParams) ([]GetThreadViewRow, error) {
	rows, err := q.query(ctx, q.getThreadViewStmt, getThreadView, arg.ID, arg.AuthorDid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetThreadViewRow
	for rows.Next() {
		var i GetThreadViewRow
		if err := rows.Scan(
			&i.ID,
			&i.Text,
			&i.ParentPostID,
			&i.RootPostID,
			&i.AuthorDid,
			&i.Handle,
			&i.CreatedAt,
			&i.HasEmbeddedMedia,
			&i.Depth,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
