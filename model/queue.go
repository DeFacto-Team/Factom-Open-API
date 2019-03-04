package model

// swagger:model
type Queue struct {
	Id   int    `json:"id"`
	Name string `json:"name" validate:"required,min=3"`
}
