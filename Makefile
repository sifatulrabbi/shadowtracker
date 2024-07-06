test-run:
	go run ./main.go -target 4000 -forward 3000 -logsdir ./logs

build:
	mkdir -p ./build
	go build -o ./build/shadowtracker ./main.go

list:
	GOPROXY=proxy.golang.org go list -m github.com/sifatulrabbi/shadowtracker@v0.1.0-beta.1

.PHONY: test-run build list
