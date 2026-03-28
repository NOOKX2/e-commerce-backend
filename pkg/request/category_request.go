package request

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required"`
	Slug string `json:"slug"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required"`
	Slug string `json:"slug"`
}

