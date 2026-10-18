package auth

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Name string `json:"name"`
}

type MeResponse struct {
	Name string `json:"name"`
}