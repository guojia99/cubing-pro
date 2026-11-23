CONFIG ?= ./local/server_local_dev.yaml

.PHONY: all help init-config api api_j api_w robot gateway build build_root admin config-test install

all: api

help:
	@printf '%s\n' \
	  'CONFIG defaults to ./local/server_local_dev.yaml; override with make TARGET CONFIG=/path/config.yaml' \
	  'init-config  Copy etc/server.yaml to CONFIG only if the destination does not exist' \
	  'api          Run API (also the default target)' \
	  'api_j        Run API with scheduled jobs (-j)' \
	  'api_w        Run API with WCA synchronization (-w; requires configured dependencies)' \
	  'gateway      Run the HTTP/TLS gateway' \
	  'robot        Run configured robots (the public example disables all instances)' \
	  'admin        Create an administrator using the existing command defaults; review before use' \
	  'build        Build ./cubing-pro (build_root remains available)' \
	  'config-test  Test only ./src/configs; does not start the application or validate services' \
	  'install      Legacy privileged build/copy/restart; does NOT install config or service units'

init-config:
	@if [ -e "$(CONFIG)" ] || [ -L "$(CONFIG)" ]; then \
	  printf 'Keeping existing config: %s\n' "$(CONFIG)"; \
	else \
	  mkdir -p "$$(dirname "$(CONFIG)")" && cp -n etc/server.yaml "$(CONFIG)"; \
	fi

api:
	go run -v main.go api --config "$(CONFIG)"

api_j:
	go run -v main.go api --config "$(CONFIG)" -j

api_w:
	go run -v main.go api --config "$(CONFIG)" -w

robot:
	go run -v main.go robot --config "$(CONFIG)"

gateway:
	go run -v main.go gateway --config "$(CONFIG)"

build: build_root

build_root:
	go build -v -o cubing-pro main.go

admin:
	go run -v main.go admin --config "$(CONFIG)"

config-test:
	go test ./src/configs

install:
	go build -v -o cubing-pro main.go
	systemctl stop cubing_pro_api.service cubing_pro_gw.service cubing_pro_robot.service
	cp cubing-pro /usr/local/bin/cubing-pro
	systemctl restart cubing_pro_api.service cubing_pro_gw.service cubing_pro_robot.service
