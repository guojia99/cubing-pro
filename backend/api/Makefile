all: build

doc:
	goctl api doc

build:
	goctl api format --dir ./api
	goctl api go --api ./api/api.api --dir . --style go_zero
	goctl api plugin -plugin goctl-swagger="swagger -filename api.json" -api  ./api/api.api -dir .
