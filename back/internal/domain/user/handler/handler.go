package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/service"
	"github.com/djsilvajr/simple-project-template/back/internal/middleware"
	"github.com/djsilvajr/simple-project-template/back/internal/response"
)

const maxBodyBytes = 1 << 20 // 1 MB

// Handler faz o papel de controller: lê a requisição, chama o service e
// escreve a resposta. Não contém regra de negócio.
type Handler struct {
	service service.Service
}

func New(svc service.Service) *Handler {
	return &Handler{service: svc}
}

// Login — POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input model.LoginInput
	if !decode(w, r, &input) {
		return
	}

	out, err := h.service.Login(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, out)
}

// Refresh — POST /auth/refresh (Authorization: Bearer <token atual, mesmo expirado>)
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	raw, ok := middleware.BearerToken(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Token de acesso ausente.")
		return
	}

	out, err := h.service.Refresh(r.Context(), raw)
	if err != nil {
		writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, out)
}

// Me — GET /auth/me
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFrom(w, r)
	if !ok {
		return
	}

	user, err := h.service.Get(r.Context(), actor, actor.ID)
	if err != nil {
		writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, user)
}

// List — GET /users?page=1&perPage=20
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFrom(w, r)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("perPage"))

	result, err := h.service.List(r.Context(), actor, model.ListUsersInput{Page: page, PerPage: perPage})
	if err != nil {
		writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, result)
}

// Get — GET /users/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFrom(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	user, err := h.service.Get(r.Context(), actor, id)
	if err != nil {
		writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, user)
}

// Create — POST /users
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFrom(w, r)
	if !ok {
		return
	}

	var input model.CreateUserInput
	if !decode(w, r, &input) {
		return
	}

	user, err := h.service.Create(r.Context(), actor, input)
	if err != nil {
		writeError(w, err)
		return
	}

	response.Success(w, http.StatusCreated, user)
}

// Update — PUT /users/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFrom(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	var input model.UpdateUserInput
	if !decode(w, r, &input) {
		return
	}

	user, err := h.service.Update(r.Context(), actor, id, input)
	if err != nil {
		writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, user)
}

// Delete — DELETE /users/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFrom(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), actor, id); err != nil {
		writeError(w, err)
		return
	}

	response.NoContent(w)
}

// --- helpers ---------------------------------------------------------------

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		response.ErrorWithDetails(w, http.StatusBadRequest, "JSON inválido.", err.Error())
		return false
	}
	return true
}

func actorFrom(w http.ResponseWriter, r *http.Request) (model.Actor, bool) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Não autenticado.")
		return model.Actor{}, false
	}
	return model.Actor{ID: claims.UserID, Role: model.Role(claims.Role)}, true
}

func pathID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(w, http.StatusBadRequest, "ID inválido.")
		return 0, false
	}
	return uint(id), true
}

// writeError converte erros de domínio em respostas HTTP.
func writeError(w http.ResponseWriter, err error) {
	if vErr, ok := usererrors.IsValidation(err); ok {
		response.ErrorWithDetails(w, http.StatusUnprocessableEntity, "Dados inválidos.", vErr.Fields)
		return
	}

	switch {
	case errors.Is(err, usererrors.ErrInvalidCredentials),
		errors.Is(err, usererrors.ErrSessionExpired):
		response.Error(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, usererrors.ErrForbidden),
		errors.Is(err, usererrors.ErrCannotDeleteSelf):
		response.Error(w, http.StatusForbidden, err.Error())
	case errors.Is(err, usererrors.ErrNotFound):
		response.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, usererrors.ErrLoginTaken),
		errors.Is(err, usererrors.ErrEmailTaken):
		response.Error(w, http.StatusConflict, err.Error())
	default:
		log.Printf("handler: erro inesperado: %v", err)
		response.Error(w, http.StatusInternalServerError, "Erro interno no servidor.")
	}
}
