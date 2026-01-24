package models

import (
	"MyBlog/modules"
)

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (u *User) RegisterUser() error {
	db := modules.GetDBManager().DB
	if err := InitUserTable(); err != nil {
		return err
	}
	query := `INSERT INTO users (name, email, password) values ($1, $2, $3)`
	_, err := db.Exec(query, u.Name, u.Email, u.Password)
	if err != nil {
		return err
	}
	return nil
}
func InitUserTable() error {
	db := modules.GetDBManager().DB
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.Exec(query)
	return err
}
