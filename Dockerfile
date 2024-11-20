FROM golang:latest

WORKDIR /usr/src/app

COPY . .

# Install dependencies
RUN apt-get update
RUN curl -Lo hugo.tar.gz "https://github.com/gohugoio/hugo/releases/download/v0.120.4/hugo_extended_0.120.4_linux-amd64.tar.gz"
RUN tar -C /usr/local/go/bin -xzf hugo.tar.gz
RUN rm -rf hugo.tar.gz
RUN chmod +x /usr/local/go/bin/hugo

# Update submodules
RUN git submodule update --init --recursive

# Build page
RUN cd web && hugo

# Build latinaapi
RUN cd ../
RUN go mod edit -dropreplace="github.com/LalatinaHub/LatinaSub-go"
# RUN go get -v github.com/LalatinaHub/LatinaSub-go@main
RUN go mod download && go mod tidy && go mod verify
RUN go build -o ./latinaapi ./cmd/latinaapi/main.go

ENV GIN_MODE=release
ENV API_MODE=true
EXPOSE 8080

CMD ["./latinaapi"]
