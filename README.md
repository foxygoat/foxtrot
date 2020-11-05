# Foxtrot

![CI](https://github.com/foxygoat/foxtrot/workflows/ci/badge.svg?branch=master)
[![PkgGoDev](https://pkg.go.dev/badge/mod/foxygo.at/foxtrot)](https://pkg.go.dev/mod/foxygo.at/foxtrot)
[![Slack chat](https://img.shields.io/badge/slack-gophers-795679?logo=slack)](https://gophers.slack.com/messages/foxygoat)

Collaborative editing is a dance.

## Step one - foxychat

Before biting off the actual Operation Transform implementation we have
decided to implement a very basic chat server and web client as it holds
a lot of the periphery tooling we will need for the ultimate goal:
Websockets, sqlite, JWT - for details see
[notes](https://docs.google.com/document/d/1p96DCIMo_0SB8OEVuh7U3jERzabJ1r-cJJLV0tLuR1U/edit#)
on google docs.

### DB

Interactively set up transient in memory DB with

    sqlite3
    .read sql/schema.sql

Alternatively use `sqlite3 out/foxtrot.db '.read sql/schema.sql'`.
