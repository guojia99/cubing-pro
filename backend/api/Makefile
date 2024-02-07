all: build

doc:
	goctl api doc

build:
	goctl api format --dir ./api
	goctl api go --api ./api/api.api --dir . --style go_zero