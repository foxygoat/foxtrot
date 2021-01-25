FROM golang:1.16beta1-buster AS builder

ARG COMMIT_SHA
ARG SEMVER
ENV COMMIT_SHA=${COMMIT_SHA}
ENV SEMVER=${SEMVER}

WORKDIR /src
COPY go.mod go.sum Makefile ./
COPY pkg pkg/
COPY cmd cmd/
RUN make install

FROM node:15.5-alpine3.10 AS frontend-builder

WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install
COPY frontend ./
RUN npm run build

FROM debian:buster-20201209
WORKDIR /
COPY --from=builder /go/bin/foxtrot /app/
COPY --from=frontend-builder /src/frontend/public /frontend/public/
CMD /app/foxtrot
