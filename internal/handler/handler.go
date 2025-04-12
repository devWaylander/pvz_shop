package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/devWaylander/pvz_store/api"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Service interface {
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

// Получение тестового токена
// (POST /dummyLogin)
func (h *Handler) PostDummyLogin(ctx context.Context, request api.PostDummyLoginRequestObject) (api.PostDummyLoginResponseObject, error) {
	return api.PostDummyLogin200JSONResponse("test"), nil
}

// Регистрация пользователя
// (POST /register)
func (h *Handler) PostRegister(ctx context.Context, req api.PostRegisterRequestObject) (api.PostRegisterResponseObject, error) {
	return api.PostRegister201JSONResponse{}, nil
}

// Авторизация пользователя
// (POST /login)
func (h *Handler) PostLogin(ctx context.Context, request api.PostLoginRequestObject) (api.PostLoginResponseObject, error) {
	return api.PostLogin200JSONResponse("s"), nil
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

// RegisterStrictHandlers регистрирует все эндпоинты strict‑сервера на chi‑роутере.
func (h *Handler) RegisterStrictHandlers(r chi.Router, sh api.ServerInterface) {
	// POST /dummyLogin
	r.Post("/dummyLogin", sh.PostDummyLogin)

	// POST /login
	r.Post("/login", sh.PostLogin)

	// POST /products
	r.Post("/products", sh.PostProducts)

	// GET /pvz
	// Для GET-эндпоинта с query параметрами создадим обёртку,
	// которая извлекает параметры из URL и передаёт в метод GetPvz.
	// Здесь для простоты предполагается, что параметры передаются как есть.
	r.Get("/pvz", func(w http.ResponseWriter, r *http.Request) {
		// Создаем переменную параметров.
		// Здесь можно использовать встроенные функции или дописать извлечение параметров более подробно.
		var params api.GetPvzParams

		// Пример: если ожидается параметр "page" в строке запроса
		if v := r.URL.Query().Get("page"); v != "" {
			// Здесь можно добавить конвертацию из строки в int
			// Например, используя strconv.Atoi
		}
		// Аналогично для других параметров: startDate, endDate, limit и т.д.

		sh.GetPvz(w, r, params)
	})

	// POST /pvz
	r.Post("/pvz", sh.PostPvz)

	// POST /pvz/{pvzId}/close_last_reception
	r.Post("/pvz/{pvzId}/close_last_reception", func(w http.ResponseWriter, r *http.Request) {
		pvzIdStr := chi.URLParam(r, "pvzId")
		// Преобразуем строку в UUID; здесь используется google/uuid,
		// предполагается, что openapi_types.UUID совместим с uuid.UUID или есть функция преобразования.
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
