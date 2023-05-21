// Code generated by sqlc. DO NOT EDIT.

package search_queries

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

func Prepare(ctx context.Context, db DBTX) (*Queries, error) {
	q := Queries{db: db}
	var err error
	if q.addAuthorStmt, err = db.PrepareContext(ctx, addAuthor); err != nil {
		return nil, fmt.Errorf("error preparing query AddAuthor: %w", err)
	}
	if q.addAuthorToClusterStmt, err = db.PrepareContext(ctx, addAuthorToCluster); err != nil {
		return nil, fmt.Errorf("error preparing query AddAuthorToCluster: %w", err)
	}
	if q.addClusterStmt, err = db.PrepareContext(ctx, addCluster); err != nil {
		return nil, fmt.Errorf("error preparing query AddCluster: %w", err)
	}
	if q.addImageStmt, err = db.PrepareContext(ctx, addImage); err != nil {
		return nil, fmt.Errorf("error preparing query AddImage: %w", err)
	}
	if q.addPostStmt, err = db.PrepareContext(ctx, addPost); err != nil {
		return nil, fmt.Errorf("error preparing query AddPost: %w", err)
	}
	if q.addPostLabelStmt, err = db.PrepareContext(ctx, addPostLabel); err != nil {
		return nil, fmt.Errorf("error preparing query AddPostLabel: %w", err)
	}
	if q.getAllUniqueLabelsStmt, err = db.PrepareContext(ctx, getAllUniqueLabels); err != nil {
		return nil, fmt.Errorf("error preparing query GetAllUniqueLabels: %w", err)
	}
	if q.getAuthorStmt, err = db.PrepareContext(ctx, getAuthor); err != nil {
		return nil, fmt.Errorf("error preparing query GetAuthor: %w", err)
	}
	if q.getAuthorStatsStmt, err = db.PrepareContext(ctx, getAuthorStats); err != nil {
		return nil, fmt.Errorf("error preparing query GetAuthorStats: %w", err)
	}
	if q.getAuthorsByHandleStmt, err = db.PrepareContext(ctx, getAuthorsByHandle); err != nil {
		return nil, fmt.Errorf("error preparing query GetAuthorsByHandle: %w", err)
	}
	if q.getClustersStmt, err = db.PrepareContext(ctx, getClusters); err != nil {
		return nil, fmt.Errorf("error preparing query GetClusters: %w", err)
	}
	if q.getImageStmt, err = db.PrepareContext(ctx, getImage); err != nil {
		return nil, fmt.Errorf("error preparing query GetImage: %w", err)
	}
	if q.getImagesForAuthorDIDStmt, err = db.PrepareContext(ctx, getImagesForAuthorDID); err != nil {
		return nil, fmt.Errorf("error preparing query GetImagesForAuthorDID: %w", err)
	}
	if q.getImagesForPostStmt, err = db.PrepareContext(ctx, getImagesForPost); err != nil {
		return nil, fmt.Errorf("error preparing query GetImagesForPost: %w", err)
	}
	if q.getMembersOfClusterStmt, err = db.PrepareContext(ctx, getMembersOfCluster); err != nil {
		return nil, fmt.Errorf("error preparing query GetMembersOfCluster: %w", err)
	}
	if q.getOldestPresentParentStmt, err = db.PrepareContext(ctx, getOldestPresentParent); err != nil {
		return nil, fmt.Errorf("error preparing query GetOldestPresentParent: %w", err)
	}
	if q.getPostStmt, err = db.PrepareContext(ctx, getPost); err != nil {
		return nil, fmt.Errorf("error preparing query GetPost: %w", err)
	}
	if q.getPostsPageByClusterAliasStmt, err = db.PrepareContext(ctx, getPostsPageByClusterAlias); err != nil {
		return nil, fmt.Errorf("error preparing query GetPostsPageByClusterAlias: %w", err)
	}
	if q.getPostsPageWithLabelStmt, err = db.PrepareContext(ctx, getPostsPageWithLabel); err != nil {
		return nil, fmt.Errorf("error preparing query GetPostsPageWithLabel: %w", err)
	}
	if q.getThreadViewStmt, err = db.PrepareContext(ctx, getThreadView); err != nil {
		return nil, fmt.Errorf("error preparing query GetThreadView: %w", err)
	}
	if q.getTopPostersStmt, err = db.PrepareContext(ctx, getTopPosters); err != nil {
		return nil, fmt.Errorf("error preparing query GetTopPosters: %w", err)
	}
	if q.getUnprocessedImagesStmt, err = db.PrepareContext(ctx, getUnprocessedImages); err != nil {
		return nil, fmt.Errorf("error preparing query GetUnprocessedImages: %w", err)
	}
	if q.updateImageStmt, err = db.PrepareContext(ctx, updateImage); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateImage: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.addAuthorStmt != nil {
		if cerr := q.addAuthorStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addAuthorStmt: %w", cerr)
		}
	}
	if q.addAuthorToClusterStmt != nil {
		if cerr := q.addAuthorToClusterStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addAuthorToClusterStmt: %w", cerr)
		}
	}
	if q.addClusterStmt != nil {
		if cerr := q.addClusterStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addClusterStmt: %w", cerr)
		}
	}
	if q.addImageStmt != nil {
		if cerr := q.addImageStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addImageStmt: %w", cerr)
		}
	}
	if q.addPostStmt != nil {
		if cerr := q.addPostStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addPostStmt: %w", cerr)
		}
	}
	if q.addPostLabelStmt != nil {
		if cerr := q.addPostLabelStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addPostLabelStmt: %w", cerr)
		}
	}
	if q.getAllUniqueLabelsStmt != nil {
		if cerr := q.getAllUniqueLabelsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getAllUniqueLabelsStmt: %w", cerr)
		}
	}
	if q.getAuthorStmt != nil {
		if cerr := q.getAuthorStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getAuthorStmt: %w", cerr)
		}
	}
	if q.getAuthorStatsStmt != nil {
		if cerr := q.getAuthorStatsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getAuthorStatsStmt: %w", cerr)
		}
	}
	if q.getAuthorsByHandleStmt != nil {
		if cerr := q.getAuthorsByHandleStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getAuthorsByHandleStmt: %w", cerr)
		}
	}
	if q.getClustersStmt != nil {
		if cerr := q.getClustersStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getClustersStmt: %w", cerr)
		}
	}
	if q.getImageStmt != nil {
		if cerr := q.getImageStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getImageStmt: %w", cerr)
		}
	}
	if q.getImagesForAuthorDIDStmt != nil {
		if cerr := q.getImagesForAuthorDIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getImagesForAuthorDIDStmt: %w", cerr)
		}
	}
	if q.getImagesForPostStmt != nil {
		if cerr := q.getImagesForPostStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getImagesForPostStmt: %w", cerr)
		}
	}
	if q.getMembersOfClusterStmt != nil {
		if cerr := q.getMembersOfClusterStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getMembersOfClusterStmt: %w", cerr)
		}
	}
	if q.getOldestPresentParentStmt != nil {
		if cerr := q.getOldestPresentParentStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getOldestPresentParentStmt: %w", cerr)
		}
	}
	if q.getPostStmt != nil {
		if cerr := q.getPostStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPostStmt: %w", cerr)
		}
	}
	if q.getPostsPageByClusterAliasStmt != nil {
		if cerr := q.getPostsPageByClusterAliasStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPostsPageByClusterAliasStmt: %w", cerr)
		}
	}
	if q.getPostsPageWithLabelStmt != nil {
		if cerr := q.getPostsPageWithLabelStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPostsPageWithLabelStmt: %w", cerr)
		}
	}
	if q.getThreadViewStmt != nil {
		if cerr := q.getThreadViewStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getThreadViewStmt: %w", cerr)
		}
	}
	if q.getTopPostersStmt != nil {
		if cerr := q.getTopPostersStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getTopPostersStmt: %w", cerr)
		}
	}
	if q.getUnprocessedImagesStmt != nil {
		if cerr := q.getUnprocessedImagesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUnprocessedImagesStmt: %w", cerr)
		}
	}
	if q.updateImageStmt != nil {
		if cerr := q.updateImageStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updateImageStmt: %w", cerr)
		}
	}
	return err
}

func (q *Queries) exec(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	case stmt != nil:
		return stmt.ExecContext(ctx, args...)
	default:
		return q.db.ExecContext(ctx, query, args...)
	}
}

func (q *Queries) query(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryContext(ctx, args...)
	default:
		return q.db.QueryContext(ctx, query, args...)
	}
}

func (q *Queries) queryRow(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) *sql.Row {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryRowContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryRowContext(ctx, args...)
	default:
		return q.db.QueryRowContext(ctx, query, args...)
	}
}

type Queries struct {
	db                             DBTX
	tx                             *sql.Tx
	addAuthorStmt                  *sql.Stmt
	addAuthorToClusterStmt         *sql.Stmt
	addClusterStmt                 *sql.Stmt
	addImageStmt                   *sql.Stmt
	addPostStmt                    *sql.Stmt
	addPostLabelStmt               *sql.Stmt
	getAllUniqueLabelsStmt         *sql.Stmt
	getAuthorStmt                  *sql.Stmt
	getAuthorStatsStmt             *sql.Stmt
	getAuthorsByHandleStmt         *sql.Stmt
	getClustersStmt                *sql.Stmt
	getImageStmt                   *sql.Stmt
	getImagesForAuthorDIDStmt      *sql.Stmt
	getImagesForPostStmt           *sql.Stmt
	getMembersOfClusterStmt        *sql.Stmt
	getOldestPresentParentStmt     *sql.Stmt
	getPostStmt                    *sql.Stmt
	getPostsPageByClusterAliasStmt *sql.Stmt
	getPostsPageWithLabelStmt      *sql.Stmt
	getThreadViewStmt              *sql.Stmt
	getTopPostersStmt              *sql.Stmt
	getUnprocessedImagesStmt       *sql.Stmt
	updateImageStmt                *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                             tx,
		tx:                             tx,
		addAuthorStmt:                  q.addAuthorStmt,
		addAuthorToClusterStmt:         q.addAuthorToClusterStmt,
		addClusterStmt:                 q.addClusterStmt,
		addImageStmt:                   q.addImageStmt,
		addPostStmt:                    q.addPostStmt,
		addPostLabelStmt:               q.addPostLabelStmt,
		getAllUniqueLabelsStmt:         q.getAllUniqueLabelsStmt,
		getAuthorStmt:                  q.getAuthorStmt,
		getAuthorStatsStmt:             q.getAuthorStatsStmt,
		getAuthorsByHandleStmt:         q.getAuthorsByHandleStmt,
		getClustersStmt:                q.getClustersStmt,
		getImageStmt:                   q.getImageStmt,
		getImagesForAuthorDIDStmt:      q.getImagesForAuthorDIDStmt,
		getImagesForPostStmt:           q.getImagesForPostStmt,
		getMembersOfClusterStmt:        q.getMembersOfClusterStmt,
		getOldestPresentParentStmt:     q.getOldestPresentParentStmt,
		getPostStmt:                    q.getPostStmt,
		getPostsPageByClusterAliasStmt: q.getPostsPageByClusterAliasStmt,
		getPostsPageWithLabelStmt:      q.getPostsPageWithLabelStmt,
		getThreadViewStmt:              q.getThreadViewStmt,
		getTopPostersStmt:              q.getTopPostersStmt,
		getUnprocessedImagesStmt:       q.getUnprocessedImagesStmt,
		updateImageStmt:                q.updateImageStmt,
	}
}
