all: api


api:
	go run -v main.go api --config ./local/server_local_dev.yaml

robot:
	go run -v main.go robot --config ./local/server_local_dev.yaml


build_root:
	go build -v -o cubing-pro main.go

admin:
	go run -v main.go admin --config ./local/server_local_dev.yaml


install:
	go build -v -o cubing-pro main.go
	systemctl stop cubing_pro_api.service cubing_pro_gw.service cubing_pro_robot.service
	cp cubing-pro /usr/local/bin/cubing-pro
	systemctl restart cubing_pro_api.service cubing_pro_gw.service cubing_pro_robot.service
