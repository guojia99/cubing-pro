all: api


api:
	go run -v cmd/root.go api --config ./etc/server_local.yaml


build_root:
	go build -v -o cubing-pro cmd/root.go



admin:
	go run -v cmd/root.go admin --config ./etc/server_local.yaml