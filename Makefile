# Include variables from the .envrc file
include .env

# ==================================================================================== # 
# Dev
# ==================================================================================== #

.PHONY: run/api
run/api:
	go run ./cmd/api

.PHONY: db/migrations/up
db/migrations/up:
	@echo 'run migration'
	migrate -path ./migrations -database ${DATABASE_URL} up

.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}



# ==================================================================================== # 
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit 
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...



## vendor: tidy and vendor dependencies
.PHONY: vendor 
vendor:
	@echo 'Tidying and verifying module dependencies...' 
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor



# ==================================================================================== # 
# BUILD
# ==================================================================================== #

## build/api: build the cmd/api application
.PHONY: dev/build/api 
dev/build/api:
	@echo 'Building dev cmd/api...'
	go build -ldflags='-s' -o=./bin/api ./cmd/api

.PHONY: prod/build/api 
prod/build/api:
	@echo 'Building prod cmd/api...'
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api

