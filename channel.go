package longpoll

import (
	"sync"
)

// Channel for sub
type Channel struct {
	topic string
	subs  map[string]*Subscriber
	sync.RWMutex
}

// NewChannel create channel
func NewChannel(topic string) *Channel {

	return &Channel{
		topic: topic,
		subs:  make(map[string]*Subscriber),
	}
}

// Topic get channel topic
func (c *Channel) Topic() string {
	return c.topic
}

func (c *Channel) setSubscriber(sub *Subscriber) {
	c.Lock()
	defer c.Unlock()

	c.subs[sub.ID()] = sub
}

func (c *Channel) getSubscriber(id string) (sub *Subscriber, ok bool) {
	c.RLock()
	defer c.RUnlock()

	sub, ok = c.subs[id]
	return
}

func (c *Channel) routeMsg(data interface{}) {
	c.Lock()
	defer c.Unlock()

	for _, sub := range c.subs {

		go func(sub *Subscriber) {
			c.purgeSub(sub)

			sub.recvData(data)

		}(sub)
	}
}

func (c *Channel) empty() bool {
	c.RLock()
	defer c.RUnlock()

	return len(c.subs) == 0
}

func (c *Channel) purgeSub(sub *Subscriber) {

	if sub.shouldPurge() {
		c.Lock()
		defer c.Unlock()

		delete(c.subs, sub.ID())
	}
}

func (c *Channel) purgeSubs() {
	c.Lock()
	defer c.Unlock()

	for _, sub := range c.subs {
		if sub.shouldPurge() {
			delete(c.subs, sub.ID())
		}
	}
}
