package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct{
	ID int64 `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Content string `json:"content"`
	Title string `json:"title"`
	UserID int64 `json:"user_id"`
	Tags []string `json:"tags"`
	Version int `json:"version"`
	Comments []Comment `json:"comments"`
	User User `json:"user"`
}

type PostWithMetadata struct{
	Post
	CommentsCount int `json:"comments_count"`
}

type PostStore struct{
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error{
	query:=`
	  insert into posts (content, title, user_id, tags)
	  values ($1, $2, $3, $4)
	  returning id, created_at, updated_at
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err:=s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err!=nil{
		return err
	}

	return nil
}

func (s *PostStore) GetByID(ctx context.Context, postID int64) (*Post, error){
	query:=`
	  select id, user_id, title, content, created_at, updated_at, tags, version
	  from posts where id=$1
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var post Post
	err:=s.db.QueryRowContext(
		ctx,
		query,
		postID,
	).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags),
		&post.Version,
	)

	if err!=nil{
		switch{
			case errors.Is(err, sql.ErrNoRows):
				return nil, ErrNotFound
			default:
				return nil, err
		}
	}

	return &post, nil
}

func (s *PostStore) DeleteByID(ctx context.Context, postID int64) error{
	query:=`delete from posts where id=$1;`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err:=s.db.ExecContext(ctx, query, postID)
	if err!=nil{
		return err
	}

	rows, err:=result.RowsAffected()
	if err!=nil{
		return err
	}

	if rows==0{
		return ErrNotFound
	}

	return nil
}

func (s *PostStore) UpdatePost(ctx context.Context, post *Post) error{
	query:=`
	  update posts
	  set title=$1, content=$2, version=version+1
	  where id=$3 and version=$4
	  returning version;
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err:=s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.ID, 
		post.Version,
	).Scan(&post.Version)

	if err!=nil{
		switch{
			case errors.Is(err, sql.ErrNoRows):
				return ErrNotFound
			default:
				return err
		}
	}

	return nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, userID int64, fq PaginatedFeedQuery) ([]PostWithMetadata, error){
	query:=`
	  select
	    p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags, u.username,
		count(c.id) as comments_count
	  from posts p
	  left join comments c on c.post_id=p.id
	  left join users u on p.user_id=u.id
	  join followers f on f.follower_id=p.user_id or p.user_id=$1
	  where 
	    f.user_id=$1 and
		(p.title ilike '%' || $4 || '%' or p.content ilike '%' || $4 || '%') and
		(p.tags @> $5::varchar[] or $5::varchar[] = '{}')
	  group by p.id, u.username
	  order by p.created_at `+ fq.Sort +`
	  limit $2 offset $3;
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err:=s.db.QueryContext(ctx, query, userID, fq.Limit, fq.Offset, fq.Search, pq.Array(fq.Tags))
	if err!=nil{
		return nil, err
	}

	defer rows.Close()

	var feed []PostWithMetadata
	for rows.Next(){
		var p PostWithMetadata
		err:=rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			&p.Version,
			pq.Array(&p.Tags),
			&p.User.Username,
			&p.CommentsCount,
		)
		if err!=nil{
			return nil, err
		}
		feed = append(feed, p)
	}

	return feed, nil
}