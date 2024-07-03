package bus

import (
	"errors"
	"time"

	trie "github.com/jangala-dev/tinygo-trie"
)

const (
	DEFAULT_Q_LEN = 10
)

var CREDS = map[string]string{
	"user":  "pass",
	"user1": "pass1",
	"user2": "pass2",
}

type Bus struct {
	trieOptions      []trie.Option
	qLength          int
	topics           *trie.Trie
	retainedMessages *trie.Trie
}

type Option func(*Bus)

func WithQLength(q int) Option {
	return func(b *Bus) {
		b.qLength = q
	}
}

func WithSingleWild(sWild string) Option {
	return func(b *Bus) {
		b.trieOptions = append(b.trieOptions, trie.WithSingleWild(sWild))
	}
}

func WithMultiWild(mWild string) Option {
	return func(b *Bus) {
		b.trieOptions = append(b.trieOptions, trie.WithMultiWild(mWild))
	}
}

func WithSeparator(sep string) Option {
	return func(b *Bus) {
		b.trieOptions = append(b.trieOptions, trie.WithSeparator(sep))
	}
}

func New(options ...Option) *Bus {
	b := &Bus{
		qLength: DEFAULT_Q_LEN,
	}
	for _, opt := range options {
		opt(b)
	}
	b.topics = trie.New(b.trieOptions...)
	b.retainedMessages = trie.New(b.trieOptions...)

	return b
}

type Message struct {
	Topic    string
	Payload  string
	Retained bool
}

type Subscription struct {
	Connection *Connection
	Topic      string
	Ch         chan Message
}

func (s *Subscription) NextMsg(timeout time.Duration) (*Message, error) {
	var msg Message
	if timeout > 0 {
		select {
		case msg = <-s.Ch:
		case <-time.After(timeout):
			return nil, errors.New("Timeout")
		}
	} else {
		msg = <-s.Ch
	}
	return &msg, nil
}

func (s *Subscription) Unsubscribe() {
	s.Connection.Unsubscribe(s.Topic, s)
}

type Connection struct {
	bus           *Bus
	subscriptions []*Subscription
}

func (c *Connection) Publish(message Message) bool {
	c.bus.Publish(message)
	return true
}

func (c *Connection) Subscribe(topic string) (*Subscription, error) {
	sub, err := c.bus.Subscribe(c, topic)
	if err != nil {
		return nil, err
	}
	c.subscriptions = append(c.subscriptions, sub)
	return sub, nil
}

func (c *Connection) Unsubscribe(topic string, subscription *Subscription) {
	c.bus.Unsubscribe(topic, subscription)
	for i, sub := range c.subscriptions {
		if sub == subscription {
			c.subscriptions = append(c.subscriptions[:i], c.subscriptions[i+1:]...)
			break
		}
	}
}

func (c *Connection) Disconnect() {
	for _, subscription := range c.subscriptions {
		c.Unsubscribe(subscription.Topic, subscription)
	}
	c.subscriptions = []*Subscription{}
}

func (b *Bus) Connect(username, password string) (*Connection, error) {
	if CREDS[username] == password {
		return &Connection{bus: b, subscriptions: []*Subscription{}}, nil
	}
	return nil, errors.New("authentication failed")
}

func (b *Bus) Subscribe(connection *Connection, topic string) (*Subscription, error) {
	topicEntryInterface, err := b.topics.Retrieve(topic)
	if err != nil {
		return nil, err
	}
	var subscriptions []*Subscription
	if topicEntryInterface == nil {
		subscriptions = make([]*Subscription, 0)
	} else {
		// Directly assert the type
		subscriptions = topicEntryInterface.([]*Subscription)
	}

	// Create the subscription
	subscription := &Subscription{
		Connection: connection,
		Topic:      topic,
		Ch:         make(chan Message, b.qLength),
	}

	subscriptions = append(subscriptions, subscription)
	b.topics.Insert(topic, subscriptions)

	// Handle retained messages
	retainedMatches := b.retainedMessages.Match(topic)
	for _, match := range retainedMatches {
		msg, ok := match.Value.(Message)
		if ok && msg.Payload != "" { // Ensure there's a payload to send
			select {
			case subscription.Ch <- msg:
			default:
				// TODO: Log that the message wasn't sent because the channel was full
			}
		}
	}

	return subscription, nil
}

func (b *Bus) Publish(message Message) {
	matches := b.topics.Match(message.Topic)
	for _, match := range matches {
		topicSubs := match.Value.([]*Subscription)
		for _, sub := range topicSubs {
			select {
			case sub.Ch <- message:
			default:
				// TODO: log this properly
			}
		}
	}
	if message.Retained {
		if message.Payload == "" {
			b.retainedMessages.Delete(message.Topic)
		} else {
			b.retainedMessages.Insert(message.Topic, message)
		}
	}
}

func (b *Bus) Unsubscribe(topic string, subscription *Subscription) {
	topicEntryInterface, _ := b.topics.Retrieve(topic) // can safely ignore error check here

	subscriptions := topicEntryInterface.([]*Subscription)
	for i, sub := range subscriptions {
		if sub == subscription {
			subscriptions = append(subscriptions[:i], subscriptions[i+1:]...)
			break
		}
	}

	if len(subscriptions) == 0 {
		b.topics.Delete(topic)
	} else {
		// Re-inserting updated subscriptions slice into the trie.
		// This might not be necessary if the slice was modified in-place,
		// but it's a good practice in case a new slice was created.
		b.topics.Insert(topic, subscriptions)
	}
}
