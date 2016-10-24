package wpool

import log "github.com/Sirupsen/logrus"

type Work interface{}

type ProcessorFn func(w Work, dispatcherName string)

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

func (d Dispatcher) Start(pfn ProcessorFn) Dispatcher {
	log.Infof("[%s] Starting [%d] workers", d.Name, d.NWorkers)

	for i := 0; i < d.NWorkers; i++ {
		worker := NewWorker(i+1, d.WorkerQueue, pfn)
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

type Worker struct {
	ID          int
	Work        chan Work
	WorkerQueue chan chan Work
	QuitChan    chan bool
	PFN         ProcessorFn
}

func NewWorker(id int, workerQueue chan chan Work, pfn ProcessorFn) Worker {

	worker := Worker{
		ID:          id,
		Work:        make(chan Work),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool),
		PFN:         pfn,
	}

	return worker
}

func (w *Worker) Start(dispatcherName string) {
	go func() {
		for {
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				w.PFN(work, dispatcherName)

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
