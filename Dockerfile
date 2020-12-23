FROM golang:1.15.5-alpine3.12 AS builder

WORKDIR /src
COPY go.mod go.sum Makefile ./
COPY pkg pkg/
COPY cmd cmd/
COPY sql sql/
RUN apk add make
RUN make install

FROM alpine:3.12.1
WORKDIR /app
COPY --from=builder /go/bin/foxtrot .
COPY static static/
CMD /app/foxtrot
