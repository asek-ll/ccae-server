
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


