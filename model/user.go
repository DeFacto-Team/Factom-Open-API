package model

type User struct {
	ID          int    `json:"id" form:"id" query:"id" validate:"required"`
	Name        string `json:"name" form:"name" query:"name" validate:"required"`
	AccessToken string `json:"access_token" form:"access_token" query:"access_token" validate:"required"`
	Usage       int    `json:"usage" form:"usage" query:"usage"`
	UsageLimit  int    `json:"usage_limit" form:"usage_limit" query:"usage_limit"`
}
