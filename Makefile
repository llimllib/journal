PWD=$(shell pwd)

journal: main.go
	go build -o journal main.go

journal-linux: main.go
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o journal-linux main.go

posts.sqlite3: parse.py
	python parse.py posts.sqlite3

deploy: journal journal-linux posts.sqlite3
	./deploy.sh

clean:
	rm -f journal*

.PHONY: deploy clean
