package inpubsub

import (
	"hash/fnv"
	"sync"
	"unsafe"
)

const (
	defaultPubSubTopicCapacity = 100000
	notificationCapacity       = 10000
	initialSubscriberCapacity  = 100
)

// PubSub is pubsub messasing object
type PubSub struct {
	mu           *sync.Mutex
	downScaleTgt chan int
	registries   registries
	publishItem  chan func() (string, interface{})
}

type registry struct {
	topic       string
	subscribers subscribers
}

type registries []*registry

type subscriber struct {
	ch        chan interface{}
	positions map[string]int
}

type subscribers []subscriber

// Subscriber is interface that wraps the Read methods
type Subscriber interface {
	Read() <-chan interface{}
}

// Read returns the channel that receive the message
func (s *subscriber) Read() <-chan interface{} {
	return s.ch
}

// NewPubSub returns PubSub object
func NewPubSub() *PubSub {
	ps := &PubSub{
		mu:           new(sync.Mutex),
		registries:   make(registries, 0, defaultPubSubTopicCapacity),
		downScaleTgt: make(chan int),
		publishItem:  make(chan func() (string, interface{}), 0),
	}

	for i := 0; i < defaultPubSubTopicCapacity; i++ {
		ps.registries = append(ps.registries, &registry{
			topic:       "",
			subscribers: make(subscribers, 0, initialSubscriberCapacity),
		})
	}

	go ps.start()

	return ps
}

// Subscribe subscribes to a topic
func (p *PubSub) Subscribe(topic string) Subscriber {
	subscriber := &subscriber{
		ch:        make(chan interface{}, notificationCapacity),
		positions: make(map[string]int),
	}

	p.mu.Lock()

	p.subscribe(topic, subscriber)

	p.mu.Unlock()

	return subscriber
}

// AddSubsrcibe adds subscribes into a topic
func (p *PubSub) AddSubsrcibe(topic string, target Subscriber) Subscriber {
	if ss, ok := target.(*subscriber); ok {

		p.mu.Lock()

		p.subscribe(topic, ss)

		p.mu.Unlock()
	}

	return target
}

func (p *PubSub) subscribe(topic string, subscriber *subscriber) {
	if topic == "" {
		return
	}

	if _, ok := subscriber.positions[topic]; ok {
		return
	}

	hash := int(generateHash(topic)) % len(p.registries)

	for i := hash; i < len(p.registries); i++ {
		registry := p.registries[i]

		if registry.topic == "" {
			registry.topic = topic
			subscriber.positions[topic] = 0
		} else if registry.topic == topic {
			subscriber.positions[topic] = len(registry.subscribers) - 1
		} else {
			continue
		}

		registry.subscribers = append(registry.subscribers, *subscriber)

		break
	}
}

func (p *PubSub) start() {
	for {
		select {
		case index := <-p.downScaleTgt:
			p.mu.Lock()

			p.registries[index].subscribers = p.downScale(p.registries[index].subscribers)

			p.mu.Unlock()
		case item := <-p.publishItem:
			p.mu.Lock()

			p.publish(item())

			p.mu.Unlock()
		}
	}
}

func (p *PubSub) downScale(targets subscribers) subscribers {
	cap := cap(targets)
	length := len(targets)

	// When length is 50% or less of cap of slice
	if cap > initialSubscriberCapacity && float32(length)/float32(cap) < 0.5 {
		aloccap := initialSubscriberCapacity
		for i := 2; length > aloccap; i++ {
			aloccap = aloccap * i
		}

		tmp := make(subscribers, 0, aloccap)

		for _, subscriber := range targets {
			tmp = append(tmp, subscriber)
		}

		return tmp
	}

	return targets
}

// Publish sends a message to subscribers subscribing to topic
func (p *PubSub) Publish(topic string, message interface{}) {
	if topic == "" {
		return
	}

	p.publishItem <- func() (string, interface{}) {
		return topic, message
	}
}

func (p *PubSub) publish(topic string, message interface{}) {
	idx := p.registries.Index(topic)
	if idx == -1 {
		return
	}

	subscribers := p.registries[idx].subscribers
	for _, subscriber := range subscribers {
		subscriber.ch <- message
	}
}

// UnSubscribe unsubscribes topic
func (p *PubSub) UnSubscribe(topic string, target Subscriber) {
	if topic == "" {
		return
	}

	if ss, ok := target.(*subscriber); ok {
		p.mu.Lock()

		p.unSubscribe(topic, ss)

		p.mu.Unlock()
	}
}

func (p *PubSub) unSubscribe(topic string, subscriber *subscriber) {
	idx := p.registries.Index(topic)
	if idx == -1 {
		return
	}

	pos := subscriber.positions[topic]
	p.registries[idx].subscribers = append(p.registries[idx].subscribers[:pos], p.registries[idx].subscribers[pos+1:]...)

	p.downScaleTgt <- idx
}

func (rs registries) Index(topic string) int {
	hash := int(generateHash(topic)) % len(rs)

	for i := hash; i < len(rs); i++ {
		if rs[i].topic == topic {
			return i
		}
	}

	return -1
}

func generateHash(text string) uint32 {
	h := fnv.New32()
	h.Write(*(*[]byte)(unsafe.Pointer(&text)))
	return h.Sum32()
}
