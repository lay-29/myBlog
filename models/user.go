package models

import (
	"MyBlog/modules"
	"fmt"
)

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func GetAllUsers() ([]User, error) {
	db := modules.GetDBManager().DB
	query := `select id,name,password,email from users`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]User, 0)
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Password, &user.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
func (u User) String() string {
	return fmt.Sprintf("User{ID:%v, Name:%v, Email:%v }", u.ID, u.Name, u.Email)
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
