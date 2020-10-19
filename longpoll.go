package longpoll

import (
	"errors"
	"os"
	"strconv"
	"sync"
	"time"
)

// errors
var (
	ErrTimeout = errors.New("polling timeout")
)

// init configs
var (
	SubPurgeInterval = time.Second
)

// ILongPoll is a contract for long poll
type ILongPoll interface {
	Subscribe(topic string, id string, timeout time.Duration) ISubscriber
	UnSubscribe(topic string, id string)
	Publish(topic string, data interface{})
	Close()
}

// LongPoll defination
type LongPoll struct {
	channels map[string]*Channel
	chQuit   chan struct{}
	sync.RWMutex
}

func parseEnv() {
	if v := os.Getenv("LP_SUB_PURGE_INTERVAL"); v != "" {
		if subPurgeIntervalInt64, err := strconv.ParseInt(v, 10, 64); err == nil {
			SubPurgeInterval = time.Duration(subPurgeIntervalInt64) * time.Second
			// unset LP_SUB_PURGE_INTERVAL to avoid side-effect for future env and env parsing
			os.Unsetenv("LP_SUB_PURGE_INTERVAL")
		}
	}

	if v := os.Getenv("LP_SUB_EVENT_BUF"); v != "" {
		if subEventBufInt64, err := strconv.ParseInt(v, 10, 64); err == nil {
			SubEventBuf = int(subEventBufInt64)
			// unset LP_SUB_EVENT_BUF to avoid side-effect for future env and env parsing
			os.Unsetenv("LP_SUB_EVENT_BUF")
		}
	}

	if v := os.Getenv("LP_SUB_LEASE_TOLERANCE"); v != "" {
		if subLeaseToleranceInt64, err := strconv.ParseInt(v, 10, 64); err == nil {
			SubLeaseTolerance = time.Duration(subLeaseToleranceInt64) * time.Second
			// unset LP_SUB_LEASE_TOLERANCE to avoid side-effect for future env and env parsing
			os.Unsetenv("LP_SUB_LEASE_TOLERANCE")
		}
	}
}

// NewLongPoll create long poll
func NewLongPoll() *LongPoll {
	parseEnv()

	lp := &LongPoll{
		channels: make(map[string]*Channel),
		chQuit:   make(chan struct{}),
	}
	go lp.purge()

	return lp
}

// Subscribe topic
func (lp *LongPoll) Subscribe(topic string, id string, timeout time.Duration) ISubscriber {

	lp.Lock()
	defer lp.Unlock()

	channel, ok := lp.channels[topic]
	if !ok {
		channel = NewChannel(topic)
		lp.channels[topic] = channel

		sub := NewSubscriber(id, channel.Topic(), timeout)
		channel.setSubscriber(sub)

		return sub
	}

	if sub, ok := channel.getSubscriber(id); ok {
		return sub
	}
	sub := NewSubscriber(id, channel.Topic(), timeout)
	channel.setSubscriber(sub)

	return sub
}

// UnSubscribe topic
func (lp *LongPoll) UnSubscribe(topic string, id string) {
	if id == "" {
		return
	}

	lp.Lock()
	defer lp.Unlock()

	channel, ok := lp.channels[topic]
	if !ok {
		return
	}

	delete(channel.subs, id)
	if channel.empty() {
		delete(lp.channels, channel.Topic())
	}
}

// Publish data to topic
func (lp *LongPoll) Publish(topic string, data interface{}) {
	lp.RLock()
	defer lp.RUnlock()

	channel, ok := lp.channels[topic]
	if !ok {
		return
	}

	channel.routeMsg(data)
}

func (lp *LongPoll) purge() {

	after := time.NewTimer(SubPurgeInterval)

	for {
		select {
		case <-lp.chQuit:
			return
		default:
		}
		lp.Lock()
		for _, channel := range lp.channels {
			channel.purgeSubs()

			if channel.empty() {
				delete(lp.channels, channel.Topic())
			}
		}
		lp.Unlock()

		after.Reset(SubPurgeInterval)
		<-after.C
	}
}

// Close long poll
func (lp *LongPoll) Close() {
	close(lp.chQuit)
}
