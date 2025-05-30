LD_FLAGS=-ldflags="-X github.com/asek-ll/aecc-server/internal/build.Time=$(shell date '+%Y-%m-%dT%H:%M:%S')"


.PHONY: run
run: build
	target/aecc-server server


.PHONY: build
build: target templ
	go build -o target/aecc-server  $(LD_FLAGS) cmd/main.go

.PHONY: templ
templ: 
	templ generate

clean:
	rm -rf target

target:
	mkdir target

.PHONY: build-amd64
build-amd64: target
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o target/aecc-server-amd64 $(LD_FLAGS) cmd/main.go 

.PHONY: build-arm64
build-arm64: target
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build -o target/aecc-server-arm64 $(LD_FLAGS) cmd/main.go

target/ccemux: target
	mkdir -p target/ccemux

target/ccemux/ccemux.json: target/ccemux
	cp test/ccemux/ccemux.json target/ccemux/ccemux.json

target/ccemux.jar: target/ccemux/ccemux.json
	curl https://github.com/asek-ll/ccemux-fork/releases/download/v0.0.2/CCEmuX-1.1.1-1.110.3-cct.jar  -o target/ccemux.jar -L

target/ccemux-plugin.jar: target/ccemux/ccemux.json
	curl https://github.com/asek-ll/ccemux-testnet-plugin/releases/download/v0.0.7/ccemux-testnet-plugin-1.0-SNAPSHOT.jar -o target/ccemux-plugin.jar -L

.PHONY: ccemux
ccemux: target/ccemux.jar target/ccemux/ccemux.json target/ccemux-plugin.jar
	java -jar target/ccemux.jar --plugin target/ccemux-plugin.jar --data-dir ./target/ccemux --computers-dir ./test/ccemux/computers/
