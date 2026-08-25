package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail=errors.New("a user with that email already exists")
	ErrDuplicateUsername=errors.New("a user with that username already exists")
)

type User struct{
	ID int64 `json:"id"`
	Username string `json:"username"`
	Email string `json:"email"`
	Password password `json:"-"`
	CreatedAt string `json:"created_at"`
	IsActive bool `json:"is_active"`
}

type password struct{
	text *string
	hash []byte
}

func (p *password) Set(text string) error{
	hash, err:=bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err!=nil{
		return err
	}

	p.text=&text
	p.hash=hash

	return nil
}

type UserStore struct{
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error{
	query:=`
	  insert into users (username, password, email)
	  values ($1, $2, $3)
	  returning id, created_at
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err:=tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password.hash,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err!=nil{
		switch{
			case err.Error()==`pq: duplicate key value violates unique constraint "users_email_key`:
				return ErrDuplicateEmail
			case err.Error()==`pq: duplicate key value violates unique constraint "users_username_key`:
				return ErrDuplicateUsername
			default:
				return err
		}
	}

	return nil
}

func (s *UserStore) GetByID(ctx context.Context, userID int64) (*User, error){
	query:=`
	  select id, username, email, password, created_at
	  from users where id=$1
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var user User
	err:=s.db.QueryRowContext(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)

	if err!=nil{
		switch{
			case errors.Is(err, sql.ErrNoRows):
				return nil, ErrNotFound
			default:
				return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error{
	return withTx(s.db, ctx, func(tx *sql.Tx) error{
		// Create a user
		if err:=s.Create(ctx, tx, user); err!=nil{
			return err
		}

		// Create user invite
		if err:=s.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err!=nil{
			return err
		}

		return nil
	})
}

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, invitationExp time.Duration, userID int64) error{
	query:=`
	  insert into user_invitations (token, user_id, expiry)
	  values ($1, $2, $3);
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	exp:=time.Now().Add(invitationExp)

	_, err:=tx.ExecContext(ctx, query, token, userID, exp)
	if err!=nil{
		return err
	}

	return nil
}

func (s *UserStore) Activate(ctx context.Context, token string) error{
	return withTx(s.db, ctx, func(tx *sql.Tx) error{
		// 1. Find the user that this token belongs to
		user, err:=s.getUserFromInvitation(ctx, tx, token)
		if err!=nil{
			return err
		}

		// 2. Update the user
		user.IsActive=true
		if err:=s.update(ctx, tx, user); err!=nil{
			return err
		}

		// 3. Clean the invitations
		if err:=s.deleteUserInvitation(ctx, tx, user.ID); err!=nil{
			return err
		}

		return nil
	})
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error){
	query:=`
	  select u.id, u.username, u.email, u.created_at, u.is_active
	  from users u
	  join user_invitations ui on u.id=ui.user_id
	  where ui.token=$1 and ui.expiry>$2;
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	hash:=sha256.Sum256([]byte(token))
	hashToken:=hex.EncodeToString(hash[:])

	user:=&User{}
	err:=tx.QueryRowContext(
		ctx,
		query,
		hashToken,
		time.Now(),
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)

	if err!=nil{
		switch err{
			case sql.ErrNoRows:
				return nil, ErrNotFound
			default:
				return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error{
	query:=`
	  update users
	  set username=$1, email=$2, is_active=$3
	  where id=$4;
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err:=tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err!=nil{
		return nil
	}

	return nil
}

func (s *UserStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userID int64) error{
	query:=`delete from user_invitations where user_id=$1;`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err:=tx.ExecContext(ctx, query, userID)
	if err!=nil{
		return err
	}

	return nil
}