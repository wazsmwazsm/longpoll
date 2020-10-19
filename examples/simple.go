package main

import (
	"fmt"
	"github.com/wazsmwazsm/longpoll"
	"log"
	"math/rand"
	"time"
)

func main() {
	lp := longpoll.NewLongPoll()

	topic := "test"
	subID := "test-id"
	subTimeout := time.Second * 5

	sub := lp.Subscribe(topic, subID, subTimeout)

	i := 0
	// simulate publisher 3~8 s send data to topic
	go func() {
		for {
			data := fmt.Sprintf("%s %d", "Hello dude", i)
			interval := time.Duration(3+rand.Intn(5)) * time.Second
			<-time.After(interval)
			lp.Publish(topic, data)
			i++
		}

	}()

	// simulate subscriber polling data with 5s timeout
	for {
		event, err := sub.WaitEvent()
		if err != nil {
			if err == longpoll.ErrTimeout {
				log.Println("End of the polling round, start new polling")
				sub = lp.Subscribe(topic, subID, subTimeout)
				continue
			}

			log.Printf("wait event err: %s", err)
		}

		log.Printf("Topic %s, subscriber %s recved data %s\n", event.Topic, event.ID, event.Data)
	}
}
