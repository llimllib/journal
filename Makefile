journal: main.go
	go build -o journal main.go

journal-linux: main.go
	# I was not able to get cross-compiling with go sqlite to work
	# https://github.com/mattn/go-sqlite3/issues/372#issuecomment-396863368
	# CC=gcc CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -o journal-linux main.go
	docker run --rm -v $(pwd):/app -w /app golang:1.14.2-stretch go build -o journal-linux

posts.sqlite3: parse.py
	python parse.py posts.sqlite3

deploy: journal journal-linux posts.sqlite3
	./deploy.sh
