FROM golang:1.19 as build

RUN mkdir /app

WORKDIR /app

COPY . /app/

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -o /go/bin/postponer

FROM alpine:3.13

# Curl для запросов healthcheck
RUN apk --no-cache add curl

COPY --from=build /go/bin/postponer /postponer

HEALTHCHECK \
    --interval=5s \
    --timeout=1s \
    --retries=3 \
    --start-period=1s \
    CMD (curl -f "http://localhost/add?queue=test.test&body=body&delay=0" || exit 1)



EXPOSE 80

CMD ["/postponer"]
