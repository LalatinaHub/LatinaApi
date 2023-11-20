FROM golang:latest

WORKDIR /usr/src/app

COPY . .

# Install dependencies
RUN go install -tags extended github.com/gohugoio/hugo@latest

# Update submodules
RUN git submodule update --init --recursive

# Build page
RUN cd web && hugo

# Build latinaapi
RUN cd ../
RUN go mod edit -dropreplace="github.com/LalatinaHub/LatinaSub-go"
RUN go get -v github.com/LalatinaHub/LatinaSub-go@main
RUN go mod download && go mod tidy && go mod verify
RUN go build -o ./latinaapi ./cmd/latinaapi/main.go

ENV GIN_MODE=release
ENV API_MODE=true
EXPOSE 8080

CMD ["./latinaapi"]
