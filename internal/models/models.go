package models

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Comment struct {
	ID        int    `json:"id"`
	Author    string `json:"author"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

type Blog struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt string    `json:"createdAt"`
	Comments  []Comment `json:"comments"`
}
