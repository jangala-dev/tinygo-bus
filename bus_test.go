package bus

import (
	"fmt"
	"testing"
	"time"
)

func TestSimplePubSub(t *testing.T) {
	b := New(WithSeparator("/"))
	conn, err := b.Connect("user", "pass")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	sub, err := conn.Subscribe("simple/topic")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	conn.Publish(Message{Topic: "simple/topic", Payload: "Hello"})

	msg, err := sub.NextMsg(time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if msg.Payload != "Hello" {
		t.Fatalf("Expected 'Hello', got '%s'", msg.Payload)
	}
}

func TestMultipleSubscribers(t *testing.T) {
	b := New(WithSeparator("/"))
	conn1, err := b.Connect("user", "pass")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	conn2, err := b.Connect("user", "pass")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	sub1, err := conn1.Subscribe("multi/topic")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	sub2, err := conn2.Subscribe("multi/topic")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	conn1.Publish(Message{Topic: "multi/topic", Payload: "Hello"})

	msg1, err := sub1.NextMsg(time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	msg2, err := sub2.NextMsg(time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if msg1.Payload != "Hello" || msg2.Payload != "Hello" {
		t.Fatalf("Expected 'Hello', got '%s' and '%s'", msg1.Payload, msg2.Payload)
	}
}

func TestMultipleTopics(t *testing.T) {
	b := New(WithSeparator("/"))
	conn, err := b.Connect("user", "pass")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	subA, err := conn.Subscribe("topic/A")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	subB, err := conn.Subscribe("topic/B")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	conn.Publish(Message{Topic: "topic/A", Payload: "MessageA"})

	msgA, err := subA.NextMsg(time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if msgA.Payload != "MessageA" {
		t.Fatalf("Expected 'MessageA', got '%s'", msgA.Payload)
	}

	msgB, err := subB.NextMsg(time.Millisecond)
	if err == nil {
		t.Fatalf("Unexpected message for topic/B: %v", msgB)
	}
	if msgB != nil {
		t.Fatalf("Expected no message for topic/B, got: %v", msgB)
	}
}

func TestCleanSubscription(t *testing.T) {
	b := New(WithSeparator("/"))
	conn, err := b.Connect("user", "pass")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Publish an old message that is not retained
	conn.Publish(Message{Topic: "clean/topic", Payload: "OldMessage"})

	// Subscribe to the topic
	sub, err := conn.Subscribe("clean/topic")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Attempt to receive a message
	msg, err := sub.NextMsg(time.Millisecond)
	if err == nil && msg != nil {
		t.Fatalf("Unexpected message: %v", msg)
	}
}

func TestConnectionCleanup(t *testing.T) {
	b := New(WithSeparator("/"))
	conn, err := b.Connect("user", "pass")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	sub, err := conn.Subscribe("cleanup/topic")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	conn.Publish(Message{Topic: "cleanup/topic", Payload: "CleanupTest"})

	msg, err := sub.NextMsg(time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if msg.Payload != "CleanupTest" {
		t.Fatalf("Expected 'CleanupTest', got '%s'", msg.Payload)
	}

	conn.Disconnect()

	// Verify the subscription is cleaned up
	topicData, err := b.topics.Retrieve("cleanup/topic")
	if err != nil {
		t.Fatalf("Failed to retrieve topic data: %v", err)
	}
	if topicData != nil {
		t.Fatalf("Subscription was not cleaned up")
	}
}

func TestRetainedMessage(t *testing.T) {
	b := New(WithSeparator("/"))
	conn, err := b.Connect("user", "pass")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	conn.Publish(Message{Topic: "retained/topic", Payload: "RetainedMessage", Retained: true})

	sub, err := conn.Subscribe("retained/topic")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	msg, err := sub.NextMsg(time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if msg.Payload != "RetainedMessage" {
		t.Fatalf("Expected 'RetainedMessage', got '%s'", msg.Payload)
	}
}

func TestUnsubscribe(t *testing.T) {
	b := New(WithSeparator("/"))
	conn, err := b.Connect("user", "pass")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	sub, err := conn.Subscribe("unsubscribe/topic")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	sub.Unsubscribe()

	conn.Publish(Message{Topic: "unsubscribe/topic", Payload: "NoReceive"})

	// Attempt to receive a message, which should not be received
	msg, err := sub.NextMsg(time.Millisecond)
	if err == nil && msg != nil {
		t.Fatalf("Unexpected message received: %v", msg)
	}
}

func TestQueueOverflow(t *testing.T) {
	b := New(WithSeparator("/"))
	conn, err := b.Connect("user", "pass")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	sub, err := conn.Subscribe("overflow/topic")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	for i := 1; i <= 11; i++ {
		conn.Publish(Message{Topic: "overflow/topic", Payload: "Message" + fmt.Sprintf("%d", i)})
	}

	for i := 1; i <= 10; i++ {
		msg, err := sub.NextMsg(time.Second)
		if err != nil {
			t.Fatalf("Failed to receive message: %v", err)
		}
		expectedPayload := "Message" + fmt.Sprintf("%d", i)
		if msg.Payload != expectedPayload {
			t.Fatalf("Expected '%s', got '%s'", expectedPayload, msg.Payload)
		}
	}

	msg, err := sub.NextMsg(time.Millisecond)
	if err == nil && msg != nil {
		t.Fatalf("Unexpected message received: %v", msg)
	}
}

func TestMultipleCredentials(t *testing.T) {
	b := New(WithSeparator("/"))
	conn1, err := b.Connect("user1", "pass1")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	sub1, err := conn1.Subscribe("multi/creds/2")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	conn2, err := b.Connect("user2", "pass2")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	sub2, err := conn2.Subscribe("multi/creds/1")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	conn1.Publish(Message{Topic: "multi/creds/1", Payload: "FromUser1"})
	conn2.Publish(Message{Topic: "multi/creds/2", Payload: "FromUser2"})

	msg1, err := sub1.NextMsg(time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	msg2, err := sub2.NextMsg(time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if msg1.Payload != "FromUser2" || msg2.Payload != "FromUser1" {
		t.Fatalf("Expected 'FromUser2' and 'FromUser1', got '%s' and '%s'", msg1.Payload, msg2.Payload)
	}
}

func TestWildcard(t *testing.T) {
	b := New(WithSeparator("/"), WithSingleWild("+"), WithMultiWild("#"))
	conn, err := b.Connect("user", "pass")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	workingSubStrings := []string{
		"wild/cards/are/fun",
		"wild/cards/are/+",
		"wild/+/are/fun",
		"wild/+/are/#",
		"wild/+/#",
		"#",
	}
	var workingSubs []*Subscription
	for _, topic := range workingSubStrings {
		sub, err := conn.Subscribe(topic)
		if err != nil {
			t.Fatalf("Failed to subscribe to topic '%s': %v", topic, err)
		}
		workingSubs = append(workingSubs, sub)
	}

	notWorkingSubStrings := []string{
		"wild/cards/are/funny",
		"wild/cards/are/+/fun",
		"wild/+/+",
		"tame/#",
	}
	var notWorkingSubs []*Subscription
	for _, topic := range notWorkingSubStrings {
		sub, err := conn.Subscribe(topic)
		if err != nil {
			t.Fatalf("Failed to subscribe to topic '%s': %v", topic, err)
		}
		notWorkingSubs = append(notWorkingSubs, sub)
	}

	conn.Publish(Message{Topic: "wild/cards/are/fun", Payload: "payload"})

	for _, sub := range workingSubs {
		msg, err := sub.NextMsg(time.Second)
		if err != nil {
			t.Fatalf("Failed to receive message: %v", err)
		}
		if msg.Payload != "payload" {
			t.Fatalf("Expected 'payload', got '%s'", msg.Payload)
		}
	}

	for _, sub := range notWorkingSubs {
		msg, err := sub.NextMsg(time.Millisecond)
		if err == nil && msg != nil {
			t.Fatalf("Unexpected message received: %v", msg)
		}
	}

	t.Log("Wildcard test passed!")
}

func TestRequestReply(t *testing.T) {
	b := New(WithSeparator("/"))

	errorChan := make(chan error, 2)

	// Helper function to simulate receiving a request and sending a reply
	startHelper := func(replyMessage string, delay time.Duration) {
		go func() {
			conn, err := b.Connect("user", "pass")
			if err != nil {
				errorChan <- err
				return
			}
			helper, err := conn.Subscribe("helpme")
			if err != nil {
				errorChan <- err
				return
			}
			rec, err := helper.NextMsg(time.Second)
			if err != nil {
				errorChan <- err
				return
			}
			time.Sleep(delay)
			conn.Publish(Message{Topic: "reply_to", Payload: replyMessage + rec.Payload})
			errorChan <- nil
		}()
	}

	// Start helpers
	startHelper("Sure ", 0)
	startHelper("No problem ", 100*time.Millisecond)

	// Wait for helpers to be ready
	time.Sleep(200 * time.Millisecond)

	// Simulate the requester
	conn, err := b.Connect("user", "pass")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	request, err := conn.Subscribe("reply_to")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	conn.Publish(Message{Topic: "helpme", Payload: "John"})

	// Check for errors from helpers
	for i := 0; i < 2; i++ {
		if err := <-errorChan; err != nil {
			t.Fatalf("Helper error: %v", err)
		}
	}

	// Check for replies
	msg, err := request.NextMsg(time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}
	if msg.Payload != "Sure John" {
		t.Fatalf("Expected 'Sure John', got '%s'", msg.Payload)
	}

	msg, err = request.NextMsg(time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}
	if msg.Payload != "No problem John" {
		t.Fatalf("Expected 'No problem John', got '%s'", msg.Payload)
	}

	t.Log("Request reply test passed!")
}
