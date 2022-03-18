
# ws-simple-pubsub

> A simple implementation of pubsub using websockets in golang

## Usage

```bash
$ make build
CGO_ENABLED=0 go build -ldflags "-X github.com/eytan-avisror/ws-simple-pubsub/pubsub.buildDate=`date +%FT%T%z` -X github.com/eytan-avisror/ws-simple-pubsub/pubsub.gitCommit=`git rev-parse HEAD`" -o bin/ws-simple-pubsub github.com/eytan-avisror/ws-simple-pubsub
chmod +x bin/ws-simple-pubsub

$ ./bin/ws-simple-pubsub --help
Usage of ./bin/ws-simple-pubsub:
  -bind-addr string
    	the address to bind the server to (default "localhost")
  -bind-port string
    	the port to bind the server to (default "8080")

$ ./bin/ws-simple-pubsub
INFO[0000] starting server on localhost:8080
```

From a websocket client, connect to the websocket endpoint at `/ws` and submit the below payloads.

### Subscribe to topic

`Subscribe` operation will subscribe to an existing topic, if the topic does not exist, it will be created and the client will be added as a subcscriber.

```json
{
    "op": "subscribe",
    "topic": "my-topic"
}
```

### Unsubscribe from topic

`Unsubscribe` operation will unsubscribe a client from a given topic.

```json
{
    "op": "unsubscribe",
    "topic": "my-topic"
}
```

### Publish to topic

`Publish` operation will publish a message to a topic, if topic does not exist, it will be created and the client will not be added as subscriber.

```json
{
    "op": "publish",
    "topic": "my-topic",
    "message": "my-message"
}
```

### Other operations

`Remove` operation will remove client from all topics.

```json
{
    "op": "remove"
}
```

`List` operation will list all available topics.

```json
{
    "op": "list"
}
```