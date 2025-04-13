package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/devWaylander/pvz_store/api"
	internalErrors "github.com/devWaylander/pvz_store/pkg/errors"
	"github.com/devWaylander/pvz_store/pkg/models"
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
	CreatePVZ(ctx context.Context, data api.PVZ) (api.PVZ, error)
	CreateReception(ctx context.Context, data api.PostReceptionsJSONBody) (api.Reception, error)
	CloseReception(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error)
	CreateProduct(ctx context.Context, data api.PostProductsJSONBody) (api.Product, error)
	DeleteLastProduct(ctx context.Context, pvzUUID uuid.UUID) error
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
	authPrincipal, err := models.GetAuthPrincipal(ctx)
	if err != nil {
		return api.PostPvz500JSONResponse{Message: err.Error()}, err
	}

	if authPrincipal.Role != string(api.Moderator) {
		err := errors.New(internalErrors.ErrForbiddenRole)
		return api.PostPvz403JSONResponse{Message: err.Error()}, nil
	}

	pvz, err := h.service.CreatePVZ(ctx, *request.Body)
	if err != nil {
		switch err.Error() {
		case internalErrors.ErrPVZExist:
			return api.PostPvz400JSONResponse{Message: err.Error()}, nil
		default:
			return api.PostPvz500JSONResponse{Message: err.Error()}, err
		}
	}

	return api.PostPvz201JSONResponse(pvz), nil
}

// Создание новой приемки товаров (только для сотрудников ПВЗ)
// (POST /receptions)
func (h *Handler) PostReceptions(ctx context.Context, request api.PostReceptionsRequestObject) (api.PostReceptionsResponseObject, error) {
	authPrincipal, err := models.GetAuthPrincipal(ctx)
	if err != nil {
		return api.PostReceptions500JSONResponse{Message: err.Error()}, err
	}

	if authPrincipal.Role != string(api.Employee) {
		err := errors.New(internalErrors.ErrForbiddenRole)
		return api.PostReceptions403JSONResponse{Message: err.Error()}, nil
	}

	reception, err := h.service.CreateReception(ctx, api.PostReceptionsJSONBody(*request.Body))
	if err != nil {
		switch err.Error() {
		case internalErrors.ErrReceptionExist, internalErrors.ErrPVZDoesntExist:
			return api.PostReceptions400JSONResponse{Message: err.Error()}, nil
		default:
			return api.PostReceptions500JSONResponse{Message: err.Error()}, err
		}
	}

	return api.PostReceptions201JSONResponse(reception), nil
}

// Добавление товара в текущую приемку (только для сотрудников ПВЗ)
// (POST /products)
func (h *Handler) PostProducts(ctx context.Context, request api.PostProductsRequestObject) (api.PostProductsResponseObject, error) {
	authPrincipal, err := models.GetAuthPrincipal(ctx)
	if err != nil {
		return api.PostProducts500JSONResponse{Message: err.Error()}, err
	}

	if authPrincipal.Role != string(api.Employee) {
		err := errors.New(internalErrors.ErrForbiddenRole)
		return api.PostProducts403JSONResponse{Message: err.Error()}, nil
	}

	product, err := h.service.CreateProduct(ctx, api.PostProductsJSONBody(*request.Body))
	if err != nil {
		switch err.Error() {
		case internalErrors.ErrPVZDoesntExist,
			internalErrors.ErrReceptionDoesntExist,
			internalErrors.ErrWrongReceptionStatus:
			return api.PostProducts400JSONResponse{Message: err.Error()}, nil
		default:
			return api.PostProducts500JSONResponse{Message: err.Error()}, err
		}
	}

	return api.PostProducts201JSONResponse(product), nil
}

// Удаление последнего добавленного товара из текущей приемки (LIFO, только для сотрудников ПВЗ)
// (POST /pvz/{pvzId}/delete_last_product)
func (h *Handler) PostPvzPvzIdDeleteLastProduct(
	ctx context.Context,
	request api.PostPvzPvzIdDeleteLastProductRequestObject) (api.PostPvzPvzIdDeleteLastProductResponseObject, error) {
	authPrincipal, err := models.GetAuthPrincipal(ctx)
	if err != nil {
		return api.PostPvzPvzIdDeleteLastProduct500JSONResponse{Message: err.Error()}, err
	}

	if authPrincipal.Role != string(api.Employee) {
		err := errors.New(internalErrors.ErrForbiddenRole)
		return api.PostPvzPvzIdDeleteLastProduct403JSONResponse{Message: err.Error()}, nil
	}

	err = h.service.DeleteLastProduct(ctx, request.PvzId)
	if err != nil {
		switch err.Error() {
		case internalErrors.ErrPVZDoesntExist,
			internalErrors.ErrReceptionDoesntExist,
			internalErrors.ErrWrongReceptionStatus,
			internalErrors.ErrNoProductsToDelete:
			return api.PostPvzPvzIdDeleteLastProduct400JSONResponse{Message: err.Error()}, nil
		default:
			return api.PostPvzPvzIdDeleteLastProduct500JSONResponse{Message: err.Error()}, err
		}
	}

	return api.PostPvzPvzIdDeleteLastProduct200Response{}, nil
}

// Закрытие последней открытой приемки товаров в рамках ПВЗ (только для сотрудников ПВЗ)
// (POST /pvz/{pvzId}/close_last_reception)
func (h *Handler) PostPvzPvzIdCloseLastReception(
	ctx context.Context,
	request api.PostPvzPvzIdCloseLastReceptionRequestObject) (api.PostPvzPvzIdCloseLastReceptionResponseObject, error) {
	authPrincipal, err := models.GetAuthPrincipal(ctx)
	if err != nil {
		return api.PostPvzPvzIdCloseLastReception500JSONResponse{Message: err.Error()}, err
	}

	if authPrincipal.Role != string(api.Employee) {
		err := errors.New(internalErrors.ErrForbiddenRole)
		return api.PostPvzPvzIdCloseLastReception403JSONResponse{Message: err.Error()}, nil
	}

	reception, err := h.service.CloseReception(ctx, request.PvzId)
	if err != nil {
		switch err.Error() {
		case internalErrors.ErrPVZDoesntExist,
			internalErrors.ErrReceptionDoesntExist,
			internalErrors.ErrWrongReceptionStatus:
			return api.PostPvzPvzIdCloseLastReception400JSONResponse{Message: err.Error()}, nil
		default:
			return api.PostPvzPvzIdCloseLastReception500JSONResponse{Message: err.Error()}, err
		}
	}

	return api.PostPvzPvzIdCloseLastReception200JSONResponse(reception), nil
}

// Получение списка ПВЗ с фильтрацией по дате приемки и пагинацией (для всех ролей)
// (GET /pvz)
func (h *Handler) GetPvz(ctx context.Context, request api.GetPvzRequestObject) (api.GetPvzResponseObject, error) {
	return api.GetPvz200JSONResponse{}, nil
}

// RegisterStrictHandlers регистрирует все эндпоинты strict‑сервера на chi‑роутере, а также занимается парсингом URL и query параметров
func (h *Handler) RegisterStrictHandlers(r chi.Router, sh api.ServerInterface) {
	// POST /dummyLogin
	r.Post("/dummyLogin", sh.PostDummyLogin)

	// POST /register
	r.Post("/register", sh.PostRegister)

	// POST /login
	r.Post("/login", sh.PostLogin)

	// POST /pvz
	r.Post("/pvz", sh.PostPvz)

	// POST /receptions
	r.Post("/receptions", sh.PostReceptions)

	// POST /products
	r.Post("/products", sh.PostProducts)

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
}
