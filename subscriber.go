package longpoll

import (
	"sync"
	"time"
)

// SubEventBuf chan buffer length
var SubEventBuf = 8 * 1024

// SubLeaseTolerance for subscriber lease
var SubLeaseTolerance time.Duration

// ISubscriber is a contract for subscriber
type ISubscriber interface {
	WaitEvent() (*SubEvent, error)
	ID() string
	Topic() string
}

// SubEvent subscribe event
type SubEvent struct {
	ID    string
	Topic string
	Data  interface{}
}

// Subscriber defination
type Subscriber struct {
	id          string
	topic       string
	chEvent     chan *SubEvent
	timeout     time.Duration
	lease       time.Duration
	refreshTime time.Duration
	sync.RWMutex
}

// NewSubscriber create Subscriber
func NewSubscriber(id string, topic string, timeout time.Duration) *Subscriber {
	sub := &Subscriber{
		id:          id,
		topic:       topic,
		chEvent:     make(chan *SubEvent, SubEventBuf),
		timeout:     timeout,
		lease:       timeout + timeout/2,
		refreshTime: time.Duration(time.Now().UnixNano()),
	}

	// avoid timeout purge before refresh
	if SubLeaseTolerance > 0 {
		sub.lease = timeout + SubLeaseTolerance
	} else {
		sub.lease = timeout + timeout/2
	}

	return sub
}

// ID get sub id
func (s *Subscriber) ID() string {
	return s.id
}

// Topic get sub topic
func (s *Subscriber) Topic() string {
	return s.topic
}

// WaitEvent block
func (s *Subscriber) WaitEvent() (*SubEvent, error) {

	defer func() { // refresh after got event data
		s.refresh()
	}()

	select {
	case data := <-s.chEvent:
		return data, nil
	case <-time.After(s.timeout):
		return nil, ErrTimeout
	}
}

func (s *Subscriber) refresh() {
	s.Lock()
	defer s.Unlock()

	s.refreshTime = time.Duration(time.Now().UnixNano())
}

// recvData and generate event
func (s *Subscriber) recvData(data interface{}) {
	s.chEvent <- &SubEvent{
		ID:    s.id,
		Topic: s.topic,
		Data:  data,
	}
}

func (s *Subscriber) shouldPurge() bool {
	s.RLock()
	defer s.RUnlock()

	now := time.Duration(time.Now().UnixNano())

	if now-s.refreshTime > s.lease {
		return true
	}

	return false
}
