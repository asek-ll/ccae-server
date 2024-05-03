
.PHONY: run
run: build
	target/aecc-server server


.PHONY: build
build: target
	go build -o target/aecc-server cmd/main.go 

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

target/ccemux-launcher.jar: target/ccemux
	curl https://emux.cc/ccemux-launcher.jar -o target/ccemux-launcher.jar

target/ccemux/ccemux.json: target/ccemux
	cp test/ccemux/ccemux.json target/ccemux/ccemux.json

.PHONY: ccemux
ccemux: target/ccemux-launcher.jar target/ccemux/ccemux.json
	java -jar target/CCEmuX-1.1.0-cct.jar --start-dir ./test/ccemux/lua --data-dir ./target/ccemux

.PHONY: cos2
cos2:
	/Applications/CraftOS-PC.app/Contents/MacOS/craftos --start-dir ./test/ccemux/lua
