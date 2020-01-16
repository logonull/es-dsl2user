daemon:
	# start our program as a background process (daemon).
	nohup go run cmd/main.go &

run:
	go run cmd/main.go

dev: run

air:
	~/.air -c .air.conf

dredd:
	dredd ./apiary.apib http://127.0.0.1:8080

testing:
	go test -v -count 1 ./...
	golangci-lint run -v

doc:
	godoc -http=:6060
