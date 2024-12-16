all: api


api:
	go run -v main.go api --config ./local/server_local_dev.yaml


robot:
	go run -v main.go robot --config ./local/server_local_dev.yaml



build_root:
	go build -v -o cubing-pro main.go

admin:
	go run -v main.go admin --config ./local/server_local_dev.yaml
