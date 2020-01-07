all:
	@echo "use release target to build release binary"

generate:
	@go generate ./...

gotest:
	go test `go list ./... | grep -v /web2600/)`
	GOOS=js GOARCH=wasm go test ./web2600/...

clean:
	@echo "removing binary and profiling files"
	@rm -f gopher2600 cpu.profile mem.profile debug.cpu.profile debug.mem.profile

build:
	go build -gcflags '-c 3 -B -+ -wb=false' 

release:
	@#go build -gcflags '-c 3 -B -+ -wb=false' -tags release
	@echo "use 'make build' for now. the release target will"
	@echo "reappear in a future commit."

profile:
	go build -gcflags '-c 3 -B -+ -wb=false' .
	./gopher2600 performance --profile roms/ROMs/Pitfall.bin
	go tool pprof -http : ./gopher2600 cpu.profile

profile_display:
	go build -gcflags '-c 3 -B -+ -wb=false' .
	./gopher2600 performance --display --profile roms/ROMs/Pitfall.bin
	go tool pprof -http : ./gopher2600 cpu.profile

web:
	cd web2600 && make release && make webserve
