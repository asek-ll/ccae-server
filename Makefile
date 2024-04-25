
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
