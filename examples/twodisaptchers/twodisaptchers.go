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
	// we start two simple dispatcher with 4 workers each
	adder := wpool.NewDispatcher("adder", 4)
	multiplier := wpool.NewDispatcher("multiplier", 4)

	// ProcessorFn "add"
	add := func(w wpool.Work, dName string, destination *wpool.Dispatcher) {
		nums := w.(Numbers)

		sum := nums.A + nums.B

		// just make this function "doing stuff"
		time.Sleep(1 * time.Second)

		// pipe resulting work to destination Dispatcher
		destination.WorkQueue <- sum
	}

	// ProcessorFn "multiply"
	multiply := func(w wpool.Work, dName string, destination *wpool.Dispatcher) {
		value := w.(int)

		// just make this function "doing stuff"
		time.Sleep(1 * time.Second)

		r := value * 2

		log.Println(dName, r)

		// pipe resulting work to destination Dispatcher
		destination.WorkQueue <- Numbers{r, 2}
	}

	// start dispatcher with 4 workers that run "add" function and pipe results to "multiplier" dispatcher
	adder.Start(add, &multiplier)

	// start dispatcher with 4 workers that run "multiply" function and pipe results to "adder" dispatcher
	multiplier.Start(multiply, &adder)

	// give "adder" some Work
	adder.WorkQueue <- Numbers{1, 2}
	adder.WorkQueue <- Numbers{2, 2}
	adder.WorkQueue <- Numbers{3, 2}
	adder.WorkQueue <- Numbers{4, 2}
	adder.WorkQueue <- Numbers{5, 2}
	adder.WorkQueue <- Numbers{6, 2}
	adder.WorkQueue <- Numbers{7, 2}
	adder.WorkQueue <- Numbers{8, 2}

	// block until app is terminated
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}
