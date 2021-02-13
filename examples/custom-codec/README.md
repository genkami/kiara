# Custom Codec

This example shows how to implement your own codec. In this example, we implement a custom codec named `WatsonCodec` that encodes messages into [WATSON](https://github.com/genkami/watson).

## Usage

First, you need to run redis-server by executing the following command:

```
$ docker-compose -f ../../docker-compose.yml up
```

then, run:

```
$ go run main.go
```

If you have `redis-cli`, you can see the fact that the message was actually encoded into WATSON.

```
$ redis-cli -p 6379
127.0.0.1:6379> subscribe room:123
Reading messages... (press Ctrl-C to quit)
1) "subscribe"
2) "room:123"
3) (integer) 1
1) "message"
2) "room:123"
3) "~?Shahaaahaha-Shahahaaaha-Shahaahahahah-Shahaahahaah-$Bubbbbbbu!Bububbububbu!Bububbbubbu!Bububbububb!Bububbubbbu!Bububbbbbu!M?Shahaaaaha-Shahaahahahah-Shahaaahaa-Shahahahaaah-$Bubbbbubb!Bububbubububu!Bubbbbb!Bububububbbu!Bububbubububu!Bubububbubbu!Bubbbbb!Bububbubbubu!Bububbububub!Bububbubububu!Bubububbububu!Bubbbbb!Bubububbubb!Bububbubbb!Bububbbbbu!Bubububbubb!Bubbbbb!Bubububbubb!Bububbubbb!Bububbubbbu!Bubububbbubu!Bubbbbb!Bububbububbu!Bububbbubbu!Bubububbbubu!Bubububbbubu!Bububbbbbu!Bububbbububu!Bububbbubbu!Bubbbbb!Bubububbububu!Bububbbbbu!Bubububbbubu!Bubbbbb!Bububbububbu!Bububbbbbu!Bubububbbub!Bubububbbubu!Bububbubbb!Bububbbbbu!Bububbububb!Bububbbubbu!Bububbbubb!Bubbbbb!Bububbubbbu!Bububbububub!Bubububbubb!Bububbubububu!Bubbbbb!Bubbubbububu!Bubbbbbbu!Bubbubbubb!Bubbubbbubu!Bubbbubububu!Bubbbububub!Bubububububu!M"
```
