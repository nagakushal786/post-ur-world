package store

import (
	"context"
	"database/sql"
)

type Comment struct{
	ID int64 `json:"id"`
	PostID int64 `json:"post_id"`
	UserID int64 `json:"user_id"`
	Content string `json:"content"`
	CreatedAt string `json:"created_at"`
	User User `json:"user"`
}

type CommentStore struct{
	db *sql.DB
}

func (s *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]Comment, error){
	query:=`
	  select c.id, c.post_id, c.user_id, c.content, c.created_at,
	  u.username, u.id from comments c join users u
	  on c.user_id=u.id where c.post_id=$1
	  order by created_at desc;
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err:=s.db.QueryContext(
		ctx,
		query,
		postID,
	)
	if err!=nil{
		return nil, err
	}

	defer rows.Close()

	comments:=[]Comment{}
	for rows.Next(){
		var c Comment
		c.User=User{}
		err:=rows.Scan(
			&c.ID,
			&c.PostID,
			&c.UserID,
			&c.Content,
			&c.CreatedAt,
			&c.User.Username,
			&c.User.ID,
		)
		if err!=nil{
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}

func (s *CommentStore) Create(ctx context.Context, comment *Comment) error{
	query:=`
	  insert into comments (post_id, user_id, content)
	  values ($1, $2, $3)
	  returning id, created_at
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err:=s.db.QueryRowContext(
		ctx,
		query,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
	)

	if err!=nil{
		return err
	}

	return nil
}