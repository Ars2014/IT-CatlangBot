package models

type Language struct {
	Name string `json:"name"`
}

type User struct {
	ID        int        `json:"id"`
	Languages []Language `json:"languages"`
}
