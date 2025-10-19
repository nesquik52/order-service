package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/stan.go"
)

func main() {
	sc, err := stan.Connect("test-cluster", "test-subscriber-123")
	if err != nil {
		log.Fatal("NATS connect failed:", err)
	}
	defer sc.Close()

	fmt.Println("Listening for messages on 'orders' channel...")

	_, err = sc.Subscribe("orders", func(msg *stan.Msg) {
		fmt.Printf("Received message: %s\n", string(msg.Data))
	})
	if err != nil {
		log.Fatal("Subscribe failed:", err)
	}

	time.Sleep(30 * time.Second)
	fmt.Println("Timeout reached")
}
