package user

import (
	"net/http"

	userusecase "admin.com/admin-api/internal/usecase/user"
)

type UserHandler struct {
	useCase userusecase.UserUseCase
}

func NewUserHandler(mux *http.ServeMux, useCase userusecase.UserUseCase) {
	handler := &UserHandler{
		useCase: useCase,
	}

	mux.HandleFunc("GET /users/{id}", handler.GetUser)
	mux.HandleFunc("GET /users", handler.GetUsers)
	mux.HandleFunc("POST /users", handler.CreateUser)
	mux.HandleFunc("PUT /users/{id}", handler.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", handler.DeleteUser)
}
