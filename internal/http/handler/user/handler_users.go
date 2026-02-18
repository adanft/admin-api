package user

import (
	"net/http"

	"admin.com/admin-api/internal/http/decoder"
	httpErrors "admin.com/admin-api/internal/http/errors"
	httprequest "admin.com/admin-api/internal/http/request"
	"admin.com/admin-api/internal/http/response"
	userusecase "admin.com/admin-api/internal/usecase/user"
	"github.com/google/uuid"
)

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, ok := userIDFromPath(w, r)
	if !ok {
		return
	}

	user, err := h.useCase.GetUser(r.Context(), id)
	if err != nil {
		writeUserBusinessError(w, r, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, response.FromUser(*user))
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req httprequest.CreateUserInput
	if err := decoder.DecodeBody(w, r, &req); err != nil {
		decoder.WriteDecodeError(w, err)
		return
	}

	user, err := h.useCase.CreateUser(r.Context(), userusecase.CreateUserInput{
		Name:     req.Name,
		LastName: req.LastName,
		Username: req.Username,
		Email:    req.Email,
		Avatar:   req.Avatar,
	})
	if err != nil {
		writeUserBusinessError(w, r, err)
		return
	}

	response.WriteSuccess(w, http.StatusCreated, response.FromUser(*user))
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.useCase.GetUsers(r.Context())
	if err != nil {
		writeUserBusinessError(w, r, err)
		return
	}

	userOutputs := make([]response.UserOutput, len(users))
	for i, user := range users {
		userOutputs[i] = response.FromUser(user)
	}

	response.WriteSuccess(w, http.StatusOK, userOutputs)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, ok := userIDFromPath(w, r)
	if !ok {
		return
	}

	if err := h.useCase.DeleteUser(r.Context(), id); err != nil {
		writeUserBusinessError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, ok := userIDFromPath(w, r)
	if !ok {
		return
	}

	var req httprequest.UpdateUserInput
	if err := decoder.DecodeBody(w, r, &req); err != nil {
		decoder.WriteDecodeError(w, err)
		return
	}

	user, err := h.useCase.UpdateUser(r.Context(), userusecase.UpdateUserInput{
		ID:       id,
		Name:     req.Name,
		LastName: req.LastName,
		Username: req.Username,
		Email:    req.Email,
		Avatar:   req.Avatar,
	})
	if err != nil {
		writeUserBusinessError(w, r, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, response.FromUser(*user))
}

func userIDFromPath(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.WriteErrorWithCode(w, httpErrors.InvalidID.Status, httpErrors.InvalidID.Code, httpErrors.InvalidID.Message)
		return uuid.Nil, false
	}

	return id, true
}
