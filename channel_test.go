package longpoll

import (
	"fmt"
	"testing"
	"time"
)

func TestNewChannel(t *testing.T) {
	channel := NewChannel("topic")

	if channel.Topic() != "topic" {
		t.Error("topic err")
	}

	if !channel.empty() {
		t.Error("subs err")
	}
}

func TestSetSub(t *testing.T) {
	channel := NewChannel("topic")

	sub := NewSubscriber("112233", channel.Topic(), time.Millisecond*100)
	channel.setSubscriber(sub)

	if len(channel.subs) != 1 {
		t.Error("subs err")
	}

	sub2 := NewSubscriber("1223", channel.Topic(), time.Millisecond*200)
	channel.setSubscriber(sub2)

	if len(channel.subs) != 2 {
		t.Error("subs err")
	}
	if cSub, ok := channel.getSubscriber(sub.ID()); sub != cSub || !ok {
		t.Error("subs err")
	}
}

func TestRouteMsg(t *testing.T) {
	channel := NewChannel("topic")

	sub := NewSubscriber("112233", channel.Topic(), time.Millisecond*100)
	channel.setSubscriber(sub)

	sub2 := NewSubscriber("1223", channel.Topic(), time.Millisecond*200)
	channel.setSubscriber(sub2)

	data := "test data"
	channel.routeMsg(data)

	event, err := sub.WaitEvent()
	if err != nil {
		t.Errorf("route msg err:%s", err)
	}

	if event.Data.(string) != data {
		t.Error("event data get err")
	}

	event2, err := sub2.WaitEvent()
	if err != nil {
		t.Errorf("route msg err:%s", err)
	}

	if event2.Data.(string) != data {
		t.Error("event data get err")
	}
}

func TestPurgeSubs(t *testing.T) {
	channel := NewChannel("topic")

	sub := NewSubscriber("112233", channel.Topic(), time.Millisecond*100)
	channel.setSubscriber(sub)

	channel.purgeSub(sub)
	if channel.empty() { // now the sub should not purge
		t.Error("purge err")
	}

	time.Sleep(time.Millisecond * 300)

	channel.purgeSub(sub)
	if !channel.empty() { // now the sub should purge
		t.Error("purge err")
	}
}

func TestPurge2(t *testing.T) {
	channel := NewChannel("topic")

	sub := NewSubscriber("112233", channel.Topic(), time.Millisecond*100)
	channel.setSubscriber(sub)

	sub2 := NewSubscriber("1223", channel.Topic(), time.Millisecond*200)
	channel.setSubscriber(sub2)

	channel.purgeSubs()
	if channel.empty() { // now the sub should not purge
		t.Error("purge err")
	}
	time.Sleep(time.Millisecond * 300)

	channel.purgeSubs()
	if !channel.empty() { // now the sub should purge
		t.Error("purge err")
	}
}

func BenchmarkRouteMsg(b *testing.B) {
	channel := NewChannel("topic")
	for i := 0; i < 1000; i++ { // 模拟 1000 个订阅者的 channel
		sub := NewSubscriber(fmt.Sprintf("%d", i), channel.Topic(), time.Second*30)
		channel.setSubscriber(sub)
	}

	data := "test data"

	for i := 0; i < b.N; i++ {
		// 启动监听
		for _, sub := range channel.subs {
			go sub.WaitEvent()
		}
		// 发送消息
		channel.routeMsg(data)
	}
}
