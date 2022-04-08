PWD=$(shell pwd)

journal: main.go
	go build -o journal main.go

journal-linux: main.go
	# I was not able to get cross-compiling with go sqlite to work
	# https://github.com/mattn/go-sqlite3/issues/372#issuecomment-396863368
	# CC=gcc-11 CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -o journal-linux main.go
	docker run --rm -v $(PWD):/app -w /app --platform=linux/amd64 golang:1.18-stretch go build -o journal-linux

posts.sqlite3: parse.py
	python parse.py posts.sqlite3

deploy: journal journal-linux posts.sqlite3
	./deploy.sh

clean:
	rm -f journal*

.PHONY: deploy clean
