FROM golang:1.16beta1-buster AS builder

WORKDIR /src
COPY go.mod go.sum Makefile ./
COPY pkg pkg/
COPY cmd cmd/
RUN make install

FROM debian:buster-20201209
WORKDIR /app
COPY --from=builder /go/bin/foxtrot .
COPY static static/
CMD /app/foxtrot
