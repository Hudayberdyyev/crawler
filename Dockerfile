FROM golang:1.16-buster AS build

WORKDIR /github.com/Hudayberdyyev/crawler
COPY . .

RUN go mod download

RUN GOOS=linux go build -o crawler cmd/main.go

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /github.com/Hudayberdyyev/crawler/crawler /crawler

CMD ./crawler