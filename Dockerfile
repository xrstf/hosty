FROM golang:1.20-alpine as builder

RUN apk add --update make git gcc musl-dev
WORKDIR /go/src/go.xrstf.de/hosty/
COPY . .
RUN make build

FROM alpine:3.17

RUN apk --no-cache add ca-certificates py-pygments
WORKDIR /app
COPY --from=builder /go/src/go.xrstf.de/hosty/hosty .
COPY --from=builder /go/src/go.xrstf.de/hosty/www www
COPY --from=builder /go/src/go.xrstf.de/hosty/resources resources
EXPOSE 80
ENTRYPOINT ["./hosty"]
