# README: Bus Module
## Overview
The Bus module provides a lightweight messaging and subscription system, using tries to provide efficient wildcard subscriptions and retained message delivery. It is built on Go.

```go
package main

import (
    "fmt"
    "time"

    "github.com/jangala-dev/tinygo-bus/bus"
)

func main() {
    // Create a new Bus
    b := bus.New(bus.WithSingleWild("+"), bus.WithMultiWild("#"), bus.WithSeparator("/"))

    // Goroutine 1: Publisher
    go func() {
        // Create a connection and publish to a topic
        conn, err := b.Connect("user", "pass")
        if err != nil {
            fmt.Println("Connection failed:", err)
            return
        }
        conn.Publish(bus.Message{Topic: "foo/bar/fizz", Payload: "Hello World!", Retained: true})
    }()

    // Goroutine 2: Subscriber
    go func() {
        // Create a connection and subscribe to a topic
        conn, err := b.Connect("user", "pass")
        if err != nil {
            fmt.Println("Connection failed:", err)
            return
        }
        sub, err := conn.Subscribe("foo/#") // using multi-level wildcard
        if err != nil {
            fmt.Println("Subscription failed:", err)
            return
        }

        // Listen for a message
        msg, err := sub.NextMsg(time.Millisecond) // 1 millisecond timeout
        if err != nil {
            fmt.Println("Failed to receive message:", err)
            return
        }
        fmt.Println("Goroutine 2:", msg.Payload)
    }()

    // Wait to ensure goroutines run
    time.Sleep(1 * time.Second)
}
```

## Features
* Efficient Storage and Retrieval: The module uses our trie module for effective message and topic storage.
* Wildcard Capabilities: The system supports single and multiple wildcard subscriptions, allowing for flexible topic matching.
* Asynchronous Message Retrieval: Subscriptions provide synchronous and asynchronous message consumption with timeouts.
* Retained Messages: The ability to keep certain messages for future delivery to subscribers.
* Basic Authentication: Basic username-password authentication before creating a connection.

## Usage
### Initialization
To create a new Bus:

```go
package main

import "github.com/jangala-dev/tinygo-bus/bus"

func main() {
    b := bus.New(
        bus.WithQLength(10),          // Optional, default is 10
        bus.WithSingleWild("+"),      // Single wildcard character, replace with your choice
        bus.WithMultiWild("#"),       // Multi-level wildcard character, replace with your choice
        bus.WithSeparator("/"),       // Separator for the topics
    )
}
```

### Connection
Create a new connection to the Bus:

```go
package main

import (
    "fmt"
    "github.com/jangala-dev/tinygo-bus/bus"
)

func main() {
    b := bus.New(bus.WithSeparator("/"))
    creds := bus.Credentials{Username: "user", Password: "pass"}
    conn, err := b.Connect(creds.Username, creds.Password)
    if err != nil {
        fmt.Println("Authentication failed!")
    }
}
```

### Publishing and Subscribing
Subscribe to a topic:

```go
sub, err := conn.Subscribe("sample/topic")
if err != nil {
    fmt.Println("Subscription failed:", err)
}
```

Retrieve the next message from a subscription:

```go
msg, err := sub.NextMsg(500 * time.Millisecond) // 0.5s timeout
if err != nil {
    fmt.Println("Timeout!")
}
```

Publish a message:

```go
conn.Publish(bus.Message{
    Topic:    "sample/topic",
    Payload:  "Hello, World!",
    Retained: false,
})
```

### Disconnecting and Cleanup
To unsubscribe from a topic:

```go
sub.Unsubscribe()
```

To disconnect a connection:

```go
conn.Disconnect()
```

## Technical Details
* The module uses goroutines for asynchronous operations.
* Subscription.NextMsg(timeout time.Duration) retrieves the next message from a subscription. If a timeout is provided, it will wait up to that duration for a message.
* Messages that carry the retained flag without a payload will result in the deletion of the retained message with the same topic.
* The system has basic authentication, using a static credentials dictionary (CREDS). This should be expanded upon for a production environment.

## Future Improvements
* Proper logging of blocked queue operations.
* Refinement of the authentication system.
* Improving the efficiency of operations with large numbers of subscriptions.
  * Federation (enabling multiple buses to interconnect)

## Setting up Private GitHub Package
To work with private repositories on GitHub, set the GOPRIVATE environment variable:

```sh
export GOPRIVATE=github.com/jangala-dev
```

## Testing
To run the tests for the Bus module, use the following command:

```sh
go test
```