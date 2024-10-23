all: api


api:
	go run -v main.go api --config ./etc/server_local.yaml


robot:
	go run -v main.go robot --config ./etc/server_local.yaml



build_root:
	go build -v -o cubing-pro main.go

admin:
	go run -v main.go admin --config ./etc/server_local.yaml
