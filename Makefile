build:
	go build -o chaosd main.go

test-integration:
	sudo go test -count=1 -p 1 -v -tags integration ./...

test-unit:
	go test -v ./...
