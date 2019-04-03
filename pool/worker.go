package pool

import (
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/service"
	log "github.com/sirupsen/logrus"
)

type Work struct {
	ID      string
	Job     *model.Chain
	Service service.Service
}

type Worker struct {
	ID            int
	WorkerChannel chan chan Work
	Channel       chan Work
	End           chan bool
}

func (w *Worker) Start() {
	go func() {
		for {
			w.WorkerChannel <- w.Channel
			select {
			case job := <-w.Channel:
				doWork(job.Job, job.Service, w.ID)
			case <-w.End:
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	log.Info("Worker ", w.ID, " stopped")
	w.End <- true
}

func doWork(chain *model.Chain, service service.Service, id int) {
	log.Info("Worker ", id, ", processing ", chain.ChainID)
	err := service.ParseAllChainEntries(chain, id)
	if err != nil {
		log.Error(err)
		service.ResetChainParsing(chain)
	}
}
