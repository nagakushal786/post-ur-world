package store

import (
	"context"
	"database/sql"
)

type Role struct{
	ID int64 `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	Level int `json:"level"`
}

type RoleStore struct{
	db *sql.DB
}

func (s *RoleStore) GetByName(ctx context.Context, roleName string) (*Role, error){
	query:=`
	  select id, name, description, level from roles
	  where name=$1;
	`

	var role Role
	err:=s.db.QueryRowContext(
		ctx,
		query,
		roleName,
	).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.Level,
	)

	if err!=nil{
		return nil, err
	}

	return &role, nil
}