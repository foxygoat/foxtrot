FROM golang:1.15.6-buster AS builder

WORKDIR /src
COPY go.mod go.sum Makefile ./
COPY pkg pkg/
COPY cmd cmd/
COPY sql sql/
RUN make install

FROM debian:buster-20201209
WORKDIR /app
COPY --from=builder /go/bin/foxtrot .
COPY static static/
CMD /app/foxtrot
