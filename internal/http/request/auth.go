package request

type RegisterInput struct {
	Name     string `json:"name"`
	LastName string `json:"lastName"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Avatar   string `json:"avatar"`
}

type LoginInput struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}
