# Foxtrot

![CI](https://github.com/foxygoat/foxtrot/workflows/ci/badge.svg?branch=master)
[![PkgGoDev](https://pkg.go.dev/badge/mod/foxygo.at/foxtrot)](https://pkg.go.dev/mod/foxygo.at/foxtrot)
[![Slack chat](https://img.shields.io/badge/slack-gophers-795679?logo=slack)](https://gophers.slack.com/messages/foxygoat)

Collaborative editing is a dance.

Before biting off the actual Operation Transform implementation we have
decided to implement a very basic chat server and web client as it holds
a lot of the periphery tooling we will need for the ultimate goal:
Websockets, sqlite, JWT - for details see
[notes](https://docs.google.com/document/d/1p96DCIMo_0SB8OEVuh7U3jERzabJ1r-cJJLV0tLuR1U/edit#)
on google docs.

## Backend

The foxtrot backend server is written in Go with a REST inspired HTTP
API for chat message history and access to other resources. Websockets
are used for new messages.

### Development

- Pre-requisites: [go 1.16](https://golang.org), [golangci-lint](https://github.com/golangci/golangci-lint/releases/tag/v1.33.2), GNU make
- Build with `make`
- View build options with `make help`

### Run foxtrot

Start the foxtrot server with a transient in-memory DB and some test
sample data with `make run`. For more options run

    make build
    out/foxtrot --help

to see for instance how to specify the path to a new or existing Sqlite
data store or define the authenticator secret.

Access the foxtrot API server locally with

    curl 'localhost:8080/api/history?room=$Kitchen'

### DB

Foxtrot uses Sqlite3 as its data store. Interactively set up transient
in memory DB with

    sqlite3
    .read pkg/foxtrot/sql/schema.sql
    .read pkg/foxtrot/sql/sample_data.sql

Alternatively, create a persistent DB with

    sqlite3 out/foxtrot.db \
      '.read pkg/foxtrot/sql/schema.sql' \
      '.read pkg/foxtrot/sql/sample_data.sql'

### Docker

On every successful merge to `master` on GitHub, the Semver patch
version number is bumped and used as git tag. Additionally
cross-platform docker images are built for linux/amd64 and linux/arm/v7
and pushed to
[foxygoat/foxtrot](https://hub.docker.com/r/foxygoat/foxtrot/tags?page=1&ordering=last_updated)
with the corresponding version tag.
