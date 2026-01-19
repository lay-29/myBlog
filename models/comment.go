package models

type Comment struct {
	ID        int    `json:"id"`
	Author    User   `json:"author"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}
