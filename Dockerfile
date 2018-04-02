FROM golang:1.10-alpine as builder

RUN apk add --update make git gcc musl-dev
WORKDIR /go/src/github.com/xrstf/hosty/
COPY . .
RUN make deps build

FROM alpine:3.7

RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /go/src/github.com/xrstf/hosty/hosty .
ENTRYPOINT ["./hosty"]