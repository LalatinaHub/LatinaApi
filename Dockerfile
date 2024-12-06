FROM oven/bun:latest AS web

WORKDIR /usr/src/web

COPY web .

RUN bun install
RUN bun run docs:build

FROM golang:latest AS app

WORKDIR /usr/src/app

COPY . .
COPY --from=web /usr/src/web/src/.vitepress/dist web/public/

# Build latinaapi
RUN go mod edit -dropreplace="github.com/LalatinaHub/LatinaSub-go"
# RUN go get -v github.com/LalatinaHub/LatinaSub-go@main
RUN go mod download && go mod tidy && go mod verify
RUN go build -o ./latinaapi ./cmd/latinaapi/main.go

ENV GIN_MODE=release
ENV API_MODE=true
EXPOSE 8080

CMD ["./latinaapi"]
