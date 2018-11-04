package main

import (
	"fmt"
	"sync"

	"github.com/rfyiamcool/inpubsub"
)

var (
	wg = &sync.WaitGroup{}
)

func main() {
	fmt.Println("start...")
	run1()
	fmt.Println("end...")
}

func run1() {
	ps := inpubsub.NewPubSub()

	subscriber1 := ps.Subscribe("t1")
	subscriber2 := ps.Subscribe("t1")
	subscriber3 := ps.Subscribe("t2")

	for _, sub := range []inpubsub.Subscriber{subscriber1, subscriber2, subscriber3} {
		go func(insub inpubsub.Subscriber) {
			for {
				got, ok := <-insub.Read()
				if !ok {
					return
				}
				fmt.Println(got)
			}
		}(sub)
	}

	ps.Publish("t1", "hi")
	ps.Publish("t2", "hello")
}
