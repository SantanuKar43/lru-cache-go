BIN_DIR := bin
APP_NAME := leru

.PHONY: build clean docker-run

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME)

docker-build:
	docker build -t $(APP_NAME):latest .

docker-run:
	docker build -t $(APP_NAME):latest .
	docker run --rm -p 9090:9090 $(APP_NAME):latest

k8s-deploy-local:
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml

k8s-delete-local:
	kubectl delete -f k8s/deployment.yaml
	kubectl delete -f k8s/service.yaml

k8s-port-forward:
	kubectl port-forward service/$(APP_NAME) 9090:9090

clean:
	rm -rf $(BIN_DIR)
