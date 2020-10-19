package longpoll

import (
	"testing"
	"time"
)

func TestNewSubscribr(t *testing.T) {
	sub := NewSubscriber("112233", "test", time.Millisecond*100)

	if sub.ID() != "112233" {
		t.Error("id gen err")
	}

	if sub.Topic() != "test" {
		t.Error("topic err")
	}
}

func TestSubTimeout(t *testing.T) {
	sub := NewSubscriber("112233", "test", time.Millisecond*100)

	if _, err := sub.WaitEvent(); err == nil || err != ErrTimeout {
		t.Error("timeout err")
	}
}

func TestSubTimeout2(t *testing.T) {
	sub := NewSubscriber("112233", "test", time.Millisecond*100)

	go func() {
		time.Sleep(time.Millisecond * 200)
		data := "test data"
		sub.recvData(data)
	}()

	if _, err := sub.WaitEvent(); err == nil || err != ErrTimeout {
		t.Error("timeout err")
	}
}
func TestSubTimeout3(t *testing.T) {
	SubLeaseTolerance = time.Millisecond * 100
	sub := NewSubscriber("112233", "test", time.Millisecond*100)

	go func() {
		time.Sleep(time.Millisecond * 200)
		data := "test data"
		sub.recvData(data)
	}()

	if _, err := sub.WaitEvent(); err == nil || err != ErrTimeout {
		t.Error("timeout err")
	}
}
func TestSubTimeout4(t *testing.T) {
	SubLeaseTolerance = time.Millisecond * 100
	sub := NewSubscriber("112233", "test", time.Millisecond*100)

	go func() {
		time.Sleep(time.Millisecond * 20)
		data := "test data"
		sub.recvData(data)
	}()

	if _, err := sub.WaitEvent(); err != nil {
		t.Error("timeout err")
	}
}
func TestSubEvent(t *testing.T) {
	sub := NewSubscriber("112233", "test", time.Millisecond*100)

	data := "test data"
	sub.recvData(data)

	event, err := sub.WaitEvent()
	if err != nil {
		t.Errorf("event get err:%s", err)
	}

	if event.ID != "112233" {
		t.Error("event id get err")
	}

	if event.Topic != "test" {
		t.Error("event id get err")
	}

	if event.Data.(string) != data {
		t.Error("event data get err")
	}
}

func TestShouldPurge(t *testing.T) {
	sub := NewSubscriber("112233", "test", time.Millisecond*100)

	if sub.shouldPurge() {
		t.Error("purge check err")
	}

	time.Sleep(time.Millisecond * 300)
	if !sub.shouldPurge() {
		t.Error("purge check err")
	}
}

func BenchmarkSubEvent(b *testing.B) {
	sub := NewSubscriber("112233", "test", time.Second*5)

	data := "test data"
	sub.recvData(data)

	for i := 0; i < b.N; i++ {

		sub.recvData(data)
		sub.WaitEvent()
	}
}
