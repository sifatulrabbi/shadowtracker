test-run:
	go run ./main.go --action start --target 4000 --forward 3000

build:
	mkdir -p ./build
	go build -o ./build/shadowtracker ./main.go

.PHONY: test-run build
