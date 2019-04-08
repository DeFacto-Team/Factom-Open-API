package model

type Link struct {
	Rel  string `json:"rel" form:"rel" query:"rel" validate:"required"`
	Href string `json:"href" form:"href" query:"href" validate:"required"`
}
