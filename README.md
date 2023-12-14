# event-bus

## Overview

This is an event bus library for usage in Go programs. It is a powerful and straightforward event management tool for Go
applications. It enables easy subscription to and publication of events on
various topics identified by string-based topics. It also allows publishers to supply data. This library is ideal for
applications requiring a simple yet effective event-driven architecture.

## Features

- **Event Subscription**: Subscribe to events using string-based topics.
- **Generic Event Data**: Publish events with generic/typed data.
- **Sequential Event Processing**: Processes published events in sequence within a single Go routine.
- **Subscription Management**: Manage subscriptions with subscriber IDs, allowing for easy unregistration.
- **Dynamic Buffer Size**: Configure the internal buffer size for event handling.

## Installation

Install the library using the following Go command:

``` bash
go get github.com/enterprizesoftware/event-bus
```

## Getting Started

### Initialization

``` go
package main

import (
	"github.com/enterprizesoftware/event-bus"
)

func main() {
    bus := eventbus.New[string]()
    defer bus.Stop()
}
```

### Subscribing to Topics

``` go
sub, err := bus.Subscribe("my-topic", func(msg string) error {
    fmt.Println("Received message:", msg)
    return nil
})
if err != nil {
    // handle subscription error
}
```

### Publishing Events

``` go
bus.Publish("my-topic", "Hello, EventBus!")
```

### Unsubscribing from Topics

``` go

// Handle - locates and removes the subscription with specified handle ID.
bus.Unsubscribe(eventbus.Handle(sub.Handle))

// Topic - removes all subscriptions for a givent topic.
bus.Unsubscribe(eventbus.Topic("my-topic"))

// Subscriber - removes all subscriptions for a given subscriber.
bus.Unsubscribe(eventbus.Subscriber(sub))

// Topic & Subscriber - removes all subscriptions for a given topic and subscriber.
bus.Unsubscribe(eventbus.Topic("my-topic"), eventbus.Subscriber(sub))
```

### Error Handling
If an error or a panic occurs during the execution of `messageHandler`, then 
the system will call an error handler if one is provided.

``` go

sub, err := bus.Subscribe("my-topic", messageHandler, eventbus.Error(myErrorHandler))
```

### Customizing EventBus

``` go
bus := eventbus.New[string](eventbus.BufferSize(200))
```

## Contributing

We welcome contributions to this project. Please submit pull requests with your proposed changes or improvements.

## License

This library is under the MIT License. See the LICENSE file in the repository for more details.