build:
	go build -o bin/ggpocker

run: build
	@./bin/ggpocker

test:
	go test -v ./...