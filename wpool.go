package wpool

import log "github.com/sirupsen/logrus"

type Work interface{}

type ProcessorFn func(w Work, dispatcherName string, destination *Dispatcher)

type Dispatcher struct {
	Name        string
	NWorkers    int
	WorkerQueue chan chan Work
	WorkQueue   chan Work
}

func NewDispatcher(name string, nworkers int) Dispatcher {
	return Dispatcher{
		Name:        name,
		NWorkers:    nworkers,
		WorkerQueue: make(chan chan Work, nworkers),
		WorkQueue:   make(chan Work, 50),
	}
}

func (d Dispatcher) Start(pfn ProcessorFn, destination *Dispatcher) Dispatcher {
	log.Infof("[%s] Starting [%d] workers", d.Name, d.NWorkers)

	for i := 0; i < d.NWorkers; i++ {
		worker := NewWorker(i+1, d.WorkerQueue, pfn, destination)
		worker.Start(d.Name)
	}

	log.Infof("[%s] Workers started", d.Name)

	go func() {
		for {
			select {
			case work := <-d.WorkQueue:
				//				log.Infof("[%s] Received work request: ", d.Name)
				go func() {
					worker := <-d.WorkerQueue

					//					log.Infof("[%s] Dispatching work request: ", d.Name)
					worker <- work
				}()
			}
		}
	}()

	return d
}

type PElement struct {
	Name     string
	NWorkers int
	PFN      ProcessorFn
}

type Pipeline struct {
	Dispatchers []Dispatcher
}

func NewPipeline(elements []PElement) Pipeline {
	pl := Pipeline{
		Dispatchers: make([]Dispatcher, len(elements)),
	}

	for index, el := range elements {
		pl.Dispatchers[index] = NewDispatcher(el.Name, el.NWorkers)
	}

	dispatcherNo := len(pl.Dispatchers)
	for index, d := range pl.Dispatchers {
		var dest *Dispatcher
		if index+1 == dispatcherNo {
			dest = nil
		} else {
			dest = &pl.Dispatchers[index+1]
		}
		d.Start(elements[index].PFN, dest)
	}
	return pl
}

type Worker struct {
	ID          int
	Work        chan Work
	WorkerQueue chan chan Work
	QuitChan    chan bool
	PFN         ProcessorFn
	Destination *Dispatcher
}

func NewWorker(id int, workerQueue chan chan Work, pfn ProcessorFn, destination *Dispatcher) Worker {

	worker := Worker{
		ID:          id,
		Work:        make(chan Work),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool),
		PFN:         pfn,
		Destination: destination,
	}

	return worker
}

func (w *Worker) Start(dispatcherName string) {
	go func() {
		for {
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				w.PFN(work, dispatcherName, w.Destination)

			case <-w.QuitChan:
				log.Infof("[%s] worker %d stopping", dispatcherName, w.ID)
				return
			}
		}
	}()
}

// stop listening for work requests (worker stops only after work is done)
func (w *Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}
