package model

// swagger:model
type Chain struct {
	Id int `json:"id"`
	Name string `json:"name" validate:"required,min=3"`
}
