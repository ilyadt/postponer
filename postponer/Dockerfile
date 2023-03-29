FROM golang:1.19 as build

RUN mkdir /app

WORKDIR /app

COPY . /app/

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    # remove debug information for decrease binary size
    # https://stackoverflow.com/a/21135705/15347300
    # 7.50 -> 5.26 MB
    -ldflags "-w -s" \
    -o /go/bin/postponer

FROM alpine:3.17

COPY --from=build /go/bin/postponer /postponer

HEALTHCHECK \
    --interval=5s \
     --timeout=1s \
     --retries=3 \ 
     CMD wget -nv -t1 --spider 'http://localhost/add?queue=test.test&body=body&delay=0' || exit 1

EXPOSE 80

CMD ["/postponer"]
