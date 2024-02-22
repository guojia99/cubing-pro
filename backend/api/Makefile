all: build


install:
	# 安装gozero相关工具
	GOPROXY=https://goproxy.cn/,direct go install github.com/zeromicro/go-zero/tools/goctl@v1.5.6
	GOPROXY=https://goproxy.cn/,direct go install github.com/zeromicro/goctl-swagger@latest

doc:
	goctl api doc

build:
	goctl api format --dir ./api
	goctl api go --api ./api/api.api --dir . --style go_zero
	goctl api plugin -plugin goctl-swagger="swagger -filename api.json" -api  ./api/api.api -dir .
