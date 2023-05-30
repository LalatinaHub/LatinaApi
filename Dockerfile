FROM node:lts as docs

WORKDIR /usr/src/web

RUN git clone https://github.com/LalatinaHub/LatinaDocs .
RUN npm install
RUN npm run build

FROM golang:latest as api

WORKDIR /usr/src/api

COPY . .

# Drop replace
RUN go mod edit -dropreplace="github.com/LalatinaHub/LatinaBot"
RUN go mod edit -dropreplace="github.com/LalatinaHub/LatinaSub-go"

RUN go get -v github.com/LalatinaHub/LatinaBot@main
RUN go get -v github.com/LalatinaHub/LatinaSub-go@main
RUN go mod download && go mod tidy && go mod verify
RUN go build -tags with_grpc,with_shadowsocksr -o ./latinaapi ./cmd/latinaapi/main.go
RUN rm -rf *

FROM golang:latest as main

WORKDIR /usr/src/app

RUN mkdir /usr/src/app/public

COPY --from=docs /usr/src/web/docs/.vitepress/dist/ /usr/src/app/public/
COPY --from=api /usr/src/api/latinaapi /usr/src/app/

ENV GIN_MODE=release
ENV API_MODE=true
EXPOSE 8080

CMD ["./latinaapi"]
