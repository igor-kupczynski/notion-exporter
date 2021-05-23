FROM golang:1.16-alpine

WORKDIR /app
COPY . .

RUN go get -d -v ./...
RUN go build -o notion-exporter -v ./...

FROM alpine:latest
RUN apk --no-cache add ca-certificates git rsync
WORKDIR /app
COPY --from=0 /app/notion-exporter .
COPY --from=0 /app/entrypoint.sh .
ENTRYPOINT ["./entrypoint.sh"]
