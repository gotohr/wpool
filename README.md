# wpool
golang worker pool library with example(s)

## usage

```go
package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gotohr/wpool"
)

type Numbers struct {
	A int
	B int
}

func main() {
	adder := wpool.NewDispatcher("adder", 4)
	multiplier := wpool.NewDispatcher("multiplier", 4)

	add := func(w wpool.Work, dName string) {
		nums := w.(Numbers)
		sum := nums.A + nums.B

		time.Sleep(1 * time.Second)

		multiplier.WorkQueue <- sum
	}

	multiply := func(w wpool.Work, dName string) {
		value := w.(int)

		time.Sleep(1 * time.Second)

		log.Println(dName, value*2)
	}

	adder.Start(add)
	multiplier.Start(multiply)

	adder.WorkQueue <- Numbers{1, 2}
	adder.WorkQueue <- Numbers{2, 2}
	adder.WorkQueue <- Numbers{3, 2}
	adder.WorkQueue <- Numbers{4, 2}
	adder.WorkQueue <- Numbers{5, 2}
	adder.WorkQueue <- Numbers{6, 2}
	adder.WorkQueue <- Numbers{7, 2}
	adder.WorkQueue <- Numbers{8, 2}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}
```
