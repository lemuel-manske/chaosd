build:
	go build -o chaosd main.go

test-integration:
	sudo go test -v -tags integration ./...

test-unit:
	go test -v ./...
