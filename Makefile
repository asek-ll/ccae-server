
.PHONY: run
run: build
	target/aecc-server server


.PHONY: build
build: target templ
	go build -o target/aecc-server cmd/main.go 

.PHONY: templ
templ: 
	templ generate

clean:
	rm -rf target

target:
	mkdir target

.PHONY: build-amd64
build-amd64: target
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o target/aecc-server-amd64 cmd/main.go 

.PHONY: build-arm64
build-arm64: target
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build -o target/aecc-server-arm64 cmd/main.go

target/ccemux: target
	mkdir -p target/ccemux

target/ccemux/ccemux.json: target/ccemux
	cp test/ccemux/ccemux.json target/ccemux/ccemux.json

target/ccemux.jar: target/ccemux/ccemux.json
	curl https://github.com/asek-ll/ccemux-fork/releases/download/v0.0.1/CCEmuX-1.1.0-1.110.3-e17ad754dd9a30130db317d9a6cefcfe56e0bfff-cct.jar -o target/ccemux.jar -L

target/ccemux-plugin.jar: target/ccemux/ccemux.json
	curl https://github.com/asek-ll/ccemux-testnet-plugin/releases/download/v0.0.3/ccemux-testnet-plugin-1.0-SNAPSHOT.jar -o target/ccemux-plugin.jar -L

.PHONY: ccemux
ccemux: target/ccemux.jar target/ccemux/ccemux.json target/ccemux-plugin.jar
	java -jar target/ccemux.jar --plugin target/ccemux-plugin.jar --start-dir ./test/ccemux/lua --data-dir ./target/ccemux
