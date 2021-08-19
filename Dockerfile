FROM golang:1.16-buster AS build

RUN go version

COPY . /github.com/Hudayberdyyev/crawler/
WORKDIR /github.com/Hudayberdyyev/crawler/

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o crawler cmd/main.go

FROM alpine:latest

WORKDIR /

COPY --from=build /github.com/Hudayberdyyev/crawler/crawler .

CMD ["./crawler"]