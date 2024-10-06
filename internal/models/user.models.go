package models

type User struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Role      int    `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	IsDeleted bool   `json:"is_deleted"`
}

type CreateUserInput struct {
	Username string `json:"username" validate:"required,min=5,max=50,alphanum"`
	Password string `json:"password" validate:"required,min=6,max=50,containsany=1234567890,containsany=QWERTYUIOPASDFGHJKLZXCVBNM"`
	Role     int    `json:"role"`
}

type UpdateUserInput struct {
	Username string `json:"username" validate:"required,min=5,max=50,alphanum"`
	Role     int    `json:"role"`
}

type UserSession struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
}

type UpdatePasswordInput struct {
	Id          int    `json:"id" validate:"required"`
	OldPassword string `json:"old_password" validate:"required"`
	Password    string `json:"password" validate:"required,min=6,max=50,containsany=1234567890,containsany=QWERTYUIOPASDFGHJKLZXCVBNM"`
}
