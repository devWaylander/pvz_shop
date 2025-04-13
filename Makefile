include .env
SERVICE = pvz
DATABASE_URL=$(DB_DATABASE_LOCAL_URL)
TEST_DATABASE_URL=$(DB_TEST_DATABASE_URL)
SWAGGER_UI_CONTAINER_NAME = swagger-ui

default: help

.PHONY: help
help: 												# Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

.PHONY: lint
lint: 												# Run linters
	@echo "🔍 Running golangci-lint..."
	@golangci-lint run --config .golangci.yaml

.PHONY: genAPI
genAPI: 										    # Generate oapi API
	oapi-codegen -generate chi-server,strict-server,types,embedded-spec -package api -o api/api.gen.go ./api/swagger.yaml

.PHONY: genGRPC
genGRPC: 										    # Generate GRPC
	protoc --go_out=. --go-grpc_out=. internal/pb/pvz_v1/pvz.proto

.PHONY: grpcCurl
grpcCurl: 										    # Проверка GRPC сервера
	grpcurl -plaintext -d '{}' localhost:3000 pvz.v1.PVZService.GetPVZList

.PHONY: installDeps
installDeps: 										# Установка необходимых зависимостей
	sudo apt install protobuf-compiler
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	sudo curl -fsSL -o /usr/local/bin/dbmate https://github.com/amacneil/dbmate/releases/latest/download/dbmate-linux-amd64
	sudo chmod +x /usr/local/bin/dbmate
	/usr/local/bin/dbmate --help

.PHONY: dropTestDB
dropTestDB: 										# Drop test database
	dbmate -u $(TEST_DATABASE_URL) drop

.PHONY: migrateTestDB
migrateTestDB: dropTestDB 							# Create database and run migrations
	dbmate -u $(TEST_DATABASE_URL) --no-dump-schema up

.PHONY: dropDB
dropDB: 										# Drop database
	dbmate -u $(DATABASE_URL) drop

.PHONY: migrateDB
migrateDB: dropDB 							# Create database and run migrations
	dbmate -u $(DATABASE_URL) --no-dump-schema up

.PHONY: stopSwaggerui
stopSwaggerui:										# Stop swaggerui
	@echo "Checking if ${SWAGGER_UI_CONTAINER_NAME} exists..."
	@CONTAINER_ID=$$(docker ps -a -q -f name=^/${SWAGGER_UI_CONTAINER_NAME}$$); \
	if [ -n "$$CONTAINER_ID" ]; then \
		echo "Stopping and removing existing container $$CONTAINER_ID..."; \
		docker stop $$CONTAINER_ID && docker rm $$CONTAINER_ID; \
	else \
		echo "No existing container to remove."; \
	fi

.PHONY: swaggerui
swaggerui: stopSwaggerui							# Run swaggerui
	@echo "Running new container ${SWAGGER_UI_CONTAINER_NAME}..."
	docker run -d --name ${SWAGGER_UI_CONTAINER_NAME} -p 5440:8080 -v ./api/swagger.yaml:/usr/share/nginx/html/swagger.yaml -v ./api/_ui/index.html:/usr/share/nginx/html/index.html -v ./api:/usr/share/nginx/html/swagger swaggerapi/swagger-ui
	@echo "Swagger UI is running on http://localhost:5440"