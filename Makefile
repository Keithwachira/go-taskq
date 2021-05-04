test:
	go test -p 1 ./...  -count=1 -v
lint:
	golangci-lint run