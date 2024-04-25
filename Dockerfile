FROM golang:1.22 as build

COPY cmd /src/cmd
COPY pkg /src/pkg
COPY internal /src/internal
COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
WORKDIR /src

ENV GOOS=linux
ENV GOARCH=arm64 
ENV CGO_ENABLED=1

RUN go build -o aecc-server cmd/main.go


FROM golang:1.22

COPY --from=build /src/aecc-server /app/server

CMD ["/app/server", "server"]
