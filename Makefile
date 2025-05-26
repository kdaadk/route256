# build-all:
# 	cd cart && GOOS=linux GOARCH=amd64 make build

# run-all: build-all
# 	docker-compose up --force-recreate --build -d



########### testing ###########
run-all-test:
	go run ./cart/cmd & \
	go run ./loms/cmd & \
	go run ./cart/mock/product

