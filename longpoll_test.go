package longpoll

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestParseEnv(t *testing.T) {

	SubPurgeInterval = time.Second
	SubEventBuf = 8 * 1024
	SubLeaseTolerance = 0

	parseEnv()

	if SubPurgeInterval != time.Second || SubEventBuf != 8*1024 || SubLeaseTolerance != 0 {
		t.Error("env parse err")
	}

	os.Setenv("LP_SUB_PURGE_INTERVAL", "aa")
	os.Setenv("LP_SUB_EVENT_BUF", "2ee")
	os.Setenv("LP_SUB_LEASE_TOLERANCE", "hello")

	parseEnv()
	if SubPurgeInterval != time.Second || SubEventBuf != 8*1024 || SubLeaseTolerance != 0 {
		t.Error("env parse err")
	}

	os.Setenv("LP_SUB_PURGE_INTERVAL", "12")
	os.Setenv("LP_SUB_EVENT_BUF", "800")
	os.Setenv("LP_SUB_LEASE_TOLERANCE", "5")

	parseEnv()

	if SubPurgeInterval != time.Second*12 || SubEventBuf != 800 || SubLeaseTolerance != time.Second*5 {
		t.Error("env parse err")
	}

	// reset
	os.Setenv("LP_SUB_PURGE_INTERVAL", "1")
	os.Setenv("LP_SUB_EVENT_BUF", "8192")
	os.Setenv("LP_SUB_LEASE_TOLERANCE", "0")

}

func TestSub(t *testing.T) {
	lp := NewLongPoll()

	sub := lp.Subscribe("test", "112233", time.Millisecond*100)

	if sub.ID() != "112233" {
		t.Error("sub id err")
	}

	if sub.Topic() != "test" {
		t.Error("topic err")
	}

	channel, ok := lp.channels["test"]
	if !ok {
		t.Error("sub channel err")
	}

	if len(channel.subs) != 1 {
		t.Error("sub channel err")
	}

	sub2 := lp.Subscribe("test", sub.ID(), time.Millisecond*100)
	if sub2 != sub {
		t.Error("sub err")
	}
}

func TestUnSub(t *testing.T) {
	lp := NewLongPoll()

	sub := lp.Subscribe("test", "112233", time.Millisecond*100)
	sub2 := lp.Subscribe("test", "2258", time.Millisecond*100)

	lp.UnSubscribe("test", "111")
	channel, ok := lp.channels["test"]
	if !ok {
		t.Error("unsub channel err")
	}

	if len(channel.subs) != 2 {
		t.Error("unsub channel err")
	}

	lp.UnSubscribe("test", "")
	channel, ok = lp.channels["test"]
	if !ok {
		t.Error("unsub channel err")
	}
	if len(channel.subs) != 2 {
		t.Error("unsub channel err")
	}

	lp.UnSubscribe("cc", "111")

	if len(channel.subs) != 2 {
		t.Error("unsub channel err")
	}

	lp.UnSubscribe("test", sub.ID())

	if len(channel.subs) != 1 {
		t.Error("unsub channel err")
	}
	lp.UnSubscribe("test", sub2.ID())

	if !channel.empty() {
		t.Error("unsub channel err")
	}

	channel, ok = lp.channels["test"]
	if ok {
		t.Error("unsub channel err")
	}
}

func TestPub(t *testing.T) {

	lp := NewLongPoll()

	sub := lp.Subscribe("test", "112233", time.Millisecond*100)
	sub2 := lp.Subscribe("test", "2258", time.Millisecond*100)

	data := "hahaha"
	lp.Publish("test", data)

	lp.Publish("bb", data)

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

func TestPubTimeout(t *testing.T) {
	lp := NewLongPoll()

	sub := lp.Subscribe("test", "112233", time.Millisecond*100)

	if _, err := sub.WaitEvent(); err == nil || err != ErrTimeout {
		t.Error("timeout err")
	}
}

func TestPubTimeout2(t *testing.T) {
	lp := NewLongPoll()

	sub := lp.Subscribe("test", "112233", time.Millisecond*100)

	go func() {
		time.Sleep(time.Millisecond * 200)
		data := "hahaha"
		lp.Publish("test", data)
	}()

	if _, err := sub.WaitEvent(); err == nil || err != ErrTimeout {
		t.Error("timeout err")
	}
}

func TestPurge(t *testing.T) {
	lp := NewLongPoll()

	lp.Subscribe("test", "112233", time.Millisecond*100)
	lp.Subscribe("test", "2258", time.Millisecond*100)
	lp.Lock()
	channel, ok := lp.channels["test"]
	if !ok {
		t.Error("purge err")
	}
	lp.Unlock()

	lp.Lock()
	if len(channel.subs) != 2 {
		t.Error("purge err")
	}
	lp.Unlock()

	time.Sleep(time.Second * 2)

	if !channel.empty() {
		t.Error("purge err")
	}

	lp.Lock()
	channel, ok = lp.channels["test"]
	if ok {
		t.Error("purge err")
	}
	lp.Unlock()

	lp.Close()
	time.Sleep(time.Second * 2)
}

func BenchmarkSubPub(b *testing.B) {
	lp := NewLongPoll()
	for i := 0; i < 1000; i++ { // 模拟一个 topic 1000 个订阅者
		lp.Subscribe("test", fmt.Sprintf("%d", i), time.Second*30)
	}

	data := "test data"

	for i := 0; i < b.N; i++ {
		// 启动监听
		for _, sub := range lp.channels["test"].subs {
			go sub.WaitEvent()
		}
		// 发送消息
		lp.Publish("test", data)
	}
}
