# Basic Usage

This example shows how to configure Kiara to use NATS as a message broker.

## Usage

First, you need to run NATS server by executing the following command:

```
$ docker-compose -f ../../docker-compose.yml up
```

then, run:

```
$ go run main.go
```
