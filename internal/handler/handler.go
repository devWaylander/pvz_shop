package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/devWaylander/pvz_store/api"
	internalErrors "github.com/devWaylander/pvz_store/pkg/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime"
)

type AuthMiddleware interface {
	DummyLogin(ctx context.Context, role api.UserRole) (api.Token, error)
	Registration(ctx context.Context, data api.PostRegisterJSONBody) (api.User, error)
	Login(ctx context.Context, data api.PostLoginJSONBody) (api.Token, error)
}

type Service interface {
}

type Handler struct {
	authMiddleware AuthMiddleware
	service        Service
}

func New(authMiddleware AuthMiddleware, service Service) *Handler {
	return &Handler{
		authMiddleware: authMiddleware,
		service:        service,
	}
}

// Получение тестового токена
// (POST /dummyLogin)
func (h *Handler) PostDummyLogin(ctx context.Context, request api.PostDummyLoginRequestObject) (api.PostDummyLoginResponseObject, error) {
	token, err := h.authMiddleware.DummyLogin(ctx, api.UserRole(request.Body.Role))
	if err != nil {
		return api.PostDummyLogin500JSONResponse{Message: err.Error()}, err
	}

	return api.PostDummyLogin200JSONResponse(token), nil
}

// Регистрация пользователя
// (POST /register)
func (h *Handler) PostRegister(ctx context.Context, req api.PostRegisterRequestObject) (api.PostRegisterResponseObject, error) {
	user, err := h.authMiddleware.Registration(ctx, api.PostRegisterJSONBody(*req.Body))
	if err != nil {
		switch err.Error() {
		case internalErrors.ErrWrongPasswordFormat:
			return api.PostRegister400JSONResponse{Message: err.Error()}, nil
		default:
			return api.PostRegister500JSONResponse{Message: err.Error()}, err
		}
	}

	return api.PostRegister201JSONResponse(user), nil
}

// Авторизация пользователя
// (POST /login)
func (h *Handler) PostLogin(ctx context.Context, request api.PostLoginRequestObject) (api.PostLoginResponseObject, error) {
	token, err := h.authMiddleware.Login(ctx, api.PostLoginJSONBody(*request.Body))
	if err != nil {
		switch err.Error() {
		case internalErrors.ErrUserNotFound:
			return api.PostLogin401JSONResponse{Message: err.Error()}, nil
		case internalErrors.ErrWrongPassword:
			return api.PostLogin401JSONResponse{Message: err.Error()}, nil
		default:
			return api.PostLogin500JSONResponse{Message: err.Error()}, err
		}
	}

	return api.PostLogin200JSONResponse(token), nil
}

// Создание ПВЗ (только для модераторов)
// (POST /pvz)
func (h *Handler) PostPvz(ctx context.Context, request api.PostPvzRequestObject) (api.PostPvzResponseObject, error) {
	return api.PostPvz201JSONResponse{}, nil
}

// Создание новой приемки товаров (только для сотрудников ПВЗ)
// (POST /receptions)
func (h *Handler) PostReceptions(ctx context.Context, request api.PostReceptionsRequestObject) (api.PostReceptionsResponseObject, error) {
	return api.PostReceptions201JSONResponse{}, nil
}

// Добавление товара в текущую приемку (только для сотрудников ПВЗ)
// (POST /products)
func (h *Handler) PostProducts(ctx context.Context, request api.PostProductsRequestObject) (api.PostProductsResponseObject, error) {
	return api.PostProducts201JSONResponse{}, nil
}

// Удаление последнего добавленного товара из текущей приемки (LIFO, только для сотрудников ПВЗ)
// (POST /pvz/{pvzId}/delete_last_product)
func (h *Handler) PostPvzPvzIdDeleteLastProduct(
	ctx context.Context,
	request api.PostPvzPvzIdDeleteLastProductRequestObject) (api.PostPvzPvzIdDeleteLastProductResponseObject, error) {
	return api.PostPvzPvzIdDeleteLastProduct200Response{}, nil
}

// Закрытие последней открытой приемки товаров в рамках ПВЗ (только для сотрудников ПВЗ)
// (POST /pvz/{pvzId}/close_last_reception)
func (h *Handler) PostPvzPvzIdCloseLastReception(
	ctx context.Context,
	request api.PostPvzPvzIdCloseLastReceptionRequestObject) (api.PostPvzPvzIdCloseLastReceptionResponseObject, error) {
	return api.PostPvzPvzIdCloseLastReception200JSONResponse{}, nil
}

// Получение списка ПВЗ с фильтрацией по дате приемки и пагинацией
// (GET /pvz)
func (h *Handler) GetPvz(ctx context.Context, request api.GetPvzRequestObject) (api.GetPvzResponseObject, error) {
	return api.GetPvz200JSONResponse{}, nil
}

// RegisterStrictHandlers регистрирует все эндпоинты strict‑сервера на chi‑роутере, а также занимается парсингом URL и query параметров
func (h *Handler) RegisterStrictHandlers(r chi.Router, sh api.ServerInterface) {
	// POST /dummyLogin
	r.Post("/dummyLogin", sh.PostDummyLogin)

	// POST /login
	r.Post("/login", sh.PostLogin)

	// POST /products
	r.Post("/products", sh.PostProducts)

	// GET /pvz
	r.Get("/pvz", func(w http.ResponseWriter, r *http.Request) {
		var params api.GetPvzParams

		err := runtime.BindQueryParameter("form", true, false, "startDate", r.URL.Query(), &params.StartDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = runtime.BindQueryParameter("form", true, false, "endDate", r.URL.Query(), &params.EndDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = runtime.BindQueryParameter("form", true, false, "page", r.URL.Query(), &params.Page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = runtime.BindQueryParameter("form", true, false, "limit", r.URL.Query(), &params.Limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sh.GetPvz(w, r, params)
	})

	// POST /pvz
	r.Post("/pvz", sh.PostPvz)

	// POST /pvz/{pvzId}/close_last_reception
	r.Post("/pvz/{pvzId}/close_last_reception", func(w http.ResponseWriter, r *http.Request) {
		pvzIdStr := chi.URLParam(r, "pvzId")
		pvzId, err := uuid.Parse(pvzIdStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid pvzId: %v", err), http.StatusBadRequest)
			return
		}

		sh.PostPvzPvzIdCloseLastReception(w, r, pvzId)
	})

	// POST /pvz/{pvzId}/delete_last_product
	r.Post("/pvz/{pvzId}/delete_last_product", func(w http.ResponseWriter, r *http.Request) {
		pvzIdStr := chi.URLParam(r, "pvzId")
		pvzId, err := uuid.Parse(pvzIdStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid pvzId: %v", err), http.StatusBadRequest)
			return
		}

		sh.PostPvzPvzIdDeleteLastProduct(w, r, pvzId)
	})

	// POST /receptions
	r.Post("/receptions", sh.PostReceptions)

	// POST /register
	r.Post("/register", sh.PostRegister)
}
