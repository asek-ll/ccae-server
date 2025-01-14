FROM alpine:3.21 AS build

RUN apk add --no-cache --update go gcc g++

WORKDIR /src
ENV GOPATH /src

COPY cmd /src/cmd
COPY pkg /src/pkg
COPY internal /src/internal
COPY go.mod /src/go.mod
COPY go.sum /src/go.sum

RUN go install github.com/a-h/templ/cmd/templ@v0.2.793
RUN /src/bin/templ generate

RUN CGO_ENABLED=1 go build -o aecc-server cmd/main.go


FROM alpine:3.21

RUN adduser -D srv

COPY --from=build --chown=srv:srv /src/aecc-server /app/server

USER srv
WORKDIR /home/srv

ENTRYPOINT ["/app/server"]
