package request

type SignUpRequest struct {
	Name     string `json:"name" validate:"required"`
	Username string `json:"username" validate:"required,min=3,max=20"`
	Phone    string `json:"phone" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
