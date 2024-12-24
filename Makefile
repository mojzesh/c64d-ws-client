test:
	@go test -v -count=1 ./...

godoc:
	@echo "Navigate to: http://localhost:6060/pkg/github.com/mojzesh/c64d-ws-client/c64dws/"
	@godoc -http=:6060 -index

install-godoc:
	@go install golang.org/x/tools/cmd/godoc@latest

run-example:
	cd examples/nested-cubes \
	&& go run .
