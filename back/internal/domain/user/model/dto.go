package model

// LoginInput é o corpo de POST /auth/login.
type LoginInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// LoginOutput é o retorno do login: token JWT + dados do usuário.
type LoginOutput struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"` // unix timestamp (segundos)
	User      *User  `json:"user"`
}

// CreateUserInput é o corpo de POST /users.
type CreateUserInput struct {
	Login    string `json:"login"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     Role   `json:"role"`
}

// UpdateUserInput é o corpo de PUT /users/{id}. Campos nil não são alterados.
type UpdateUserInput struct {
	Login    *string `json:"login"`
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	Role     *Role   `json:"role"`
}

// ListUsersInput representa os parâmetros de paginação de GET /users.
type ListUsersInput struct {
	Page    int
	PerPage int
}
