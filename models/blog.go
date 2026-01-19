package models

type Blog struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Author    User      `json:"author"`
	Content   string    `json:"content"`
	CreatedAt string    `json:"createdAt"`
	Comments  []Comment `json:"comments"`
}
