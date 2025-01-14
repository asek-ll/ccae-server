FROM golang:1.22 AS build

COPY cmd /src/cmd
COPY pkg /src/pkg
COPY internal /src/internal
COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
WORKDIR /src

RUN go install github.com/a-h/templ/cmd/templ@v0.2.793
RUN templ generate

ENV CGO_ENABLED=1

RUN go build -o aecc-server cmd/main.go


FROM alpine:3.21

RUN apk add libc6-compat

COPY --from=build /src/aecc-server /app/server

CMD ["/app/server", "server"]
