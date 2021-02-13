# Chat

This is a chat application built with Kiara and [gorilla/websocket](https://github.com/gorilla/websocket/).

![demo chat application](../../doc/img/kiara-chat-demo.gif)

## Usage

First, you need to run redis-server by executing the following command:

```
$ docker-compose -f ../../docker-compose.yml up
```

then, run:

```
$ go run main.go
```
