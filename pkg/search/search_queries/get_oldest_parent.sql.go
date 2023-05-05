// Code generated by sqlc. DO NOT EDIT.
// source: get_oldest_parent.sql

package search_queries

import (
	"context"
	"database/sql"
	"time"
)

const getOldestPresentParent = `-- name: GetOldestPresentParent :one
WITH RECURSIVE cte AS (
    SELECT 
        id, text, parent_post_id, root_post_id, author_did, created_at, has_embedded_media, parent_relationship
    FROM 
        posts 
    WHERE 
        posts.id = $1
    UNION ALL
    SELECT 
        p.id, p.text, p.parent_post_id, p.root_post_id, p.author_did, p.created_at, p.has_embedded_media, p.parent_relationship
    FROM 
        posts p
    INNER JOIN 
        cte ON p.id = cte.parent_post_id
)
SELECT 
    id, text, parent_post_id, root_post_id, author_did, created_at, has_embedded_media, parent_relationship
FROM 
    cte
WHERE 
    parent_post_id IS NULL
`

type GetOldestPresentParentRow struct {
	ID                 string         `json:"id"`
	Text               string         `json:"text"`
	ParentPostID       sql.NullString `json:"parent_post_id"`
	RootPostID         sql.NullString `json:"root_post_id"`
	AuthorDid          string         `json:"author_did"`
	CreatedAt          time.Time      `json:"created_at"`
	HasEmbeddedMedia   bool           `json:"has_embedded_media"`
	ParentRelationship sql.NullString `json:"parent_relationship"`
}

func (q *Queries) GetOldestPresentParent(ctx context.Context, id string) (GetOldestPresentParentRow, error) {
	row := q.queryRow(ctx, q.getOldestPresentParentStmt, getOldestPresentParent, id)
	var i GetOldestPresentParentRow
	err := row.Scan(
		&i.ID,
		&i.Text,
		&i.ParentPostID,
		&i.RootPostID,
		&i.AuthorDid,
		&i.CreatedAt,
		&i.HasEmbeddedMedia,
		&i.ParentRelationship,
	)
	return i, err
}
