package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/devWaylander/pvz_store/api"
	"github.com/devWaylander/pvz_store/config"
	"github.com/devWaylander/pvz_store/internal/handler"
	"github.com/devWaylander/pvz_store/internal/middleware/auth"
	"github.com/devWaylander/pvz_store/internal/middleware/cors"
	"github.com/devWaylander/pvz_store/internal/middleware/logger"
	"github.com/devWaylander/pvz_store/internal/repo"
	"github.com/devWaylander/pvz_store/internal/service"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type E2eIntegrationTestSuite struct {
	suite.Suite
	dbPool *sqlx.DB
}

func TestE2eIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(E2eIntegrationTestSuite))
}

func (s *E2eIntegrationTestSuite) SetupSuite() {
	cfg, err := config.Parse()
	if err != nil {
		s.T().Fatalf("failed to parse config: %v", err)
	}

	// Подключение к БД
	db, err := sqlx.Connect("postgres", cfg.DB.DBTestUrl)
	if err != nil {
		s.T().Fatalf("failed to connect to db: %v", err)
	}
	db.SetMaxOpenConns(cfg.DB.DBMaxConnections)
	db.SetMaxIdleConns(cfg.DB.DBMaxConnections)
	db.SetConnMaxLifetime(cfg.DB.DBLifeTimeConnection)
	db.SetConnMaxIdleTime(cfg.DB.DBMaxConnIdleTime)
	s.dbPool = db

	// Подготовка зависимостей
	repoInstance := repo.New(db)
	authRepo := auth.NewRepo(db)
	serviceInstance := service.New(repoInstance)
	authMiddleware := auth.NewMiddleware(authRepo, cfg.Common.JWTSecret)

	// chi router + middleware + handler
	r := chi.NewRouter()
	r.Use(logger.Middleware())
	r.Use(cors.Middleware())
	r.Use(authMiddleware.AuthContextEnrichingMiddleware)

	swagger, err := api.GetSwagger()
	if err != nil {
		s.T().Fatalf("failed to get swagger: %v", err)
	}
	r.Use(nethttpmiddleware.OapiRequestValidatorWithOptions(swagger, &nethttpmiddleware.Options{
		SilenceServersWarning: true,
		Options: openapi3filter.Options{
			AuthenticationFunc: authMiddleware.Middleware(),
		},
	}))

	handlerInstance := handler.New(authMiddleware, serviceInstance)
	strictHandler := api.NewStrictHandler(handlerInstance, nil)
	handlerInstance.RegisterStrictHandlers(r, strictHandler)

	// Запуск HTTP сервера в горутине
	go func() {
		err := http.ListenAndServe(":"+cfg.Common.Port, r)
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("failed to start server: %v", err))
		}
	}()
}

func (s *E2eIntegrationTestSuite) TestFullIntegrationFlow() {
	t := s.T()
	client := HttpClient{}
	newUuid := uuid.New()
	newRegDate := time.Now()

	// dummy login
	loginBody := api.PostDummyLoginJSONBody{
		Role: api.PostDummyLoginJSONBodyRole(api.Moderator),
	}
	reqBody, err := json.Marshal(loginBody)
	require.NoError(t, err)
	resp, respBody, err := client.SendJsonReq("", http.MethodPost, BaseURL+"/dummyLogin", reqBody)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// moderator
	var moderatorToken api.Token
	err = json.Unmarshal(respBody, &moderatorToken)
	require.NoError(t, err)
	require.Greater(t, len(moderatorToken), 0)

	loginBody = api.PostDummyLoginJSONBody{
		Role: api.PostDummyLoginJSONBodyRole(api.Employee),
	}
	reqBody, err = json.Marshal(loginBody)
	require.NoError(t, err)

	resp, respBody, err = client.SendJsonReq("", http.MethodPost, BaseURL+"/dummyLogin", reqBody)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// employee
	var employeeToken api.Token
	err = json.Unmarshal(respBody, &employeeToken)
	require.NoError(t, err)
	require.Greater(t, len(moderatorToken), 0)

	// Create PVZ
	createPVZBody := api.PostPvzJSONRequestBody{
		Id:               &newUuid,
		City:             "Москва",
		RegistrationDate: &newRegDate,
	}

	body, err := json.Marshal(createPVZBody)
	require.NoError(t, err)

	resp, _, err = client.SendJsonReq(moderatorToken, http.MethodPost, BaseURL+"/pvz", body)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Create Reception
	createReceptionBody := api.PostReceptionsJSONRequestBody{
		PvzId: newUuid,
	}
	body, err = json.Marshal(createReceptionBody)
	require.NoError(t, err)

	resp, respBody, err = client.SendJsonReq(employeeToken, http.MethodPost, BaseURL+"/receptions", body)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var reception api.Reception
	err = json.Unmarshal(respBody, &reception)
	require.NoError(t, err)

	// Add 50 products
	for i := 0; i < 30; i++ {
		createProductBody := api.PostProductsJSONRequestBody{
			PvzId: *createPVZBody.Id,
			Type:  api.PostProductsJSONBodyType(api.ProductTypeЭлектроника),
		}
		body, err := json.Marshal(createProductBody)
		require.NoError(t, err)

		resp, _, err = client.SendJsonReq(employeeToken, http.MethodPost, BaseURL+"/products", body)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}
	for i := 0; i < 10; i++ {
		createProductBody := api.PostProductsJSONRequestBody{
			PvzId: *createPVZBody.Id,
			Type:  api.PostProductsJSONBodyType(api.ProductTypeОдежда),
		}
		body, err := json.Marshal(createProductBody)
		require.NoError(t, err)

		resp, _, err = client.SendJsonReq(employeeToken, http.MethodPost, BaseURL+"/products", body)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}
	for i := 0; i < 10; i++ {
		createProductBody := api.PostProductsJSONRequestBody{
			PvzId: *createPVZBody.Id,
			Type:  api.PostProductsJSONBodyType(api.ProductTypeОбувь),
		}
		body, err := json.Marshal(createProductBody)
		require.NoError(t, err)

		resp, _, err = client.SendJsonReq(employeeToken, http.MethodPost, BaseURL+"/products", body)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Close reception
	url := fmt.Sprintf(BaseURL+"/pvz/%s/close_last_reception", createReceptionBody.PvzId)
	resp, _, err = client.SendJsonReq(employeeToken, http.MethodPost, url, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
