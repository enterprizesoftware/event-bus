package eventbus

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// EventBus

type eventBusInfo struct {
	mtx    *sync.Mutex
	once   *sync.Once
	done   chan struct{}
	buffer int
}

type EventBus[T any] struct {
	eventBusInfo
	notifications chan notification[T]
	topics        map[string][]Subscription[T]
}

func New[T any](opts ...Option) *EventBus[T] {
	ebi := eventBusInfo{
		mtx:    &sync.Mutex{},
		once:   &sync.Once{},
		done:   make(chan struct{}),
		buffer: 100,
	}

	for _, opt := range opts {
		opt(&ebi)
	}

	eb := &EventBus[T]{
		eventBusInfo:  ebi,
		notifications: make(chan notification[T], ebi.buffer),
		topics:        make(map[string][]Subscription[T]),
	}

	eb.Start()

	return eb
}

type Option func(*eventBusInfo)

func BufferSize(size int) Option {
	return func(ebi *eventBusInfo) {
		ebi.buffer = size
	}
}

func (b *EventBus[T]) Start() *EventBus[T] {
	b.once.Do(func() {
		go func() {
			for {
				select {
				case <-b.done:
					return
				case n := <-b.notifications:
					b.notify(n)
				}
			}
		}()
	})
	return b
}

func (b *EventBus[T]) Stop() {
	close(b.done)
}

func (b *EventBus[T]) Publish(topic string, data T) {
	b.notifications <- notification[T]{topic: topic, data: data}
}

func (b *EventBus[T]) Subscribe(topic string, handler func(T) error, opts ...SubscriptionOption) (Subscription[T], error) {
	s := Subscription[T]{
		subscriptionInfo: subscriptionInfo{
			topic: topic,
		},
		handler: handler,
	}

	for _, opt := range opts {
		opt(&s.subscriptionInfo)
	}

	if s.topic == "" {
		return s, fmt.Errorf("topic cannot be empty")
	}

	if s.handler == nil {
		return s, fmt.Errorf("handler cannot be nil")
	}

	s.handle = nextHandle()

	b.mtx.Lock()
	defer b.mtx.Unlock()

	subs := b.topics[s.topic]

	for i, sub := range subs {
		if sub.subscriber == s.subscriber {
			if !s.replace {
				return s, fmt.Errorf("subscriber already exists for topic: %s", s.topic)
			}
			subs[i] = s
			return s, nil
		}
	}

	subs = append(subs, s)
	b.topics[s.topic] = subs

	return s, nil

}

func (b *EventBus[T]) notify(n notification[T]) {
	var s *Subscription[T]

	b.mtx.Lock()
	defer b.mtx.Unlock()

	defer func() {
		if r := recover(); r != nil {
			if s != nil && s.errorFn != nil {
				s.errorFn(fmt.Errorf("panic: %v", r))
			}
		}
	}()

	subs := b.topics[n.topic]

	for _, sub := range subs {
		s = &sub
		err := s.handler(n.data)
		if err != nil {
			if s.errorFn != nil {
				s.errorFn(err)
			}
		}
	}

	s = nil
}

func (b *EventBus[T]) Unsubscribe(opts ...SubscriptionOption) {
	si := &subscriptionInfo{}

	for _, opt := range opts {
		opt(si)
	}

	if si.handle > 0 {
		b.removeHandle(si.handle)
		return
	}

	if si.subscriber != "" {
		if si.topic != "" {
			b.removeTopicSubscriber(si.topic, si.subscriber)
			return
		}
		b.removeSubscriber(si.subscriber)
		return
	}

	if si.topic != "" {
		b.removeTopic(si.topic)
		return
	}
}

func (b *EventBus[T]) removeHandle(handle int64) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	for topic, subs := range b.topics {
		for i, sub := range subs {
			if sub.handle == handle {
				subs = append(subs[:i], subs[i+1:]...)
				b.topics[topic] = subs
			}
		}
	}
}

func (b *EventBus[T]) removeTopicSubscriber(topic string, subscriber string) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	subs := b.topics[topic]

	for i, sub := range subs {
		if sub.subscriber == subscriber {
			subs = append(subs[:i], subs[i+1:]...)
			b.topics[topic] = subs
		}
	}
}

func (b *EventBus[T]) removeSubscriber(subscriber string) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	for topic, subs := range b.topics {
		for i, sub := range subs {
			if sub.subscriber == subscriber {
				subs = append(subs[:i], subs[i+1:]...)
				b.topics[topic] = subs
			}
		}
	}
}

func (b *EventBus[T]) removeTopic(topic string) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	delete(b.topics, topic)
}

type notification[T any] struct {
	topic string
	data  T
}

type SubscriptionOption func(builder *subscriptionInfo)

func Topic(topic string) SubscriptionOption {
	return func(sb *subscriptionInfo) {
		sb.topic = topic
	}
}

func Subscriber(id string) SubscriptionOption {
	return func(sb *subscriptionInfo) {
		sb.subscriber = id
	}
}

func Error(fn func(error)) SubscriptionOption {
	return func(sb *subscriptionInfo) {
		sb.errorFn = fn
	}
}

func Handle(handle int64) SubscriptionOption {
	return func(sb *subscriptionInfo) {
		sb.handle = handle
	}
}

func Replace() SubscriptionOption {
	return func(sb *subscriptionInfo) {
		sb.replace = true
	}
}

type Subscription[T any] struct {
	subscriptionInfo
	handler func(T) error
}

type subscriptionInfo struct {
	topic      string
	subscriber string
	handle     int64
	errorFn    func(error)
	replace    bool
}

var handleCounter int64 = 0

func nextHandle() int64 {
	return atomic.AddInt64(&handleCounter, 1)
}
