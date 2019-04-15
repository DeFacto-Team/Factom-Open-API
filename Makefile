init:
	if ! hash dep 2>/dev/null; then go get -u github.com/golang/dep/cmd/dep; fi
	if ! hash mockgen 2>/dev/null; then go get github.com/golang/mock/mockgen; fi
	if ! hash swagger 2>/dev/null; then go get -u github.com/go-swagger/go-swagger/cmd/swagger; fi
	dep ensure

run: db-run
	dep ensure
	go run main.go

test: db-run
	dep ensure
	go test -v ./test/
	$(MAKE) db-stop

bench: db-run
	go test -bench=. -benchmem ./test/
	$(MAKE) db-stop

spec:
	swagger generate spec -m -o ./spec/api.json

spec-ui: spec
	swagger serve -F=swagger ./spec/api.json

db-run: db-stop
	 docker run -d -p 5433:5432 --name factom-open-api-db postgres
	 sleep 3

db-stop:
	docker container stop factom-open-api-db >/dev/null 2>&1 || exit 0

.PHONY: run test db-run db-stop spec spec-ui