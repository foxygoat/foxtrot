FROM golang:1.14-alpine3.12 as builder
WORKDIR /src
COPY go.mod go.sum main.go ./
RUN go install

FROM alpine:3.12.1
COPY --from=builder /go/bin/foxtrot /
COPY static /static
ENTRYPOINT ["/foxtrot"]
EXPOSE 8080
