package main

import (
	log "github.com/sirupsen/logrus"
)

type MRResponse struct {
	status string
	err    error
}

type Scheduler struct {
	requests     chan MergeRequest
	responses    chan MRResponse
	status       chan WorkerStatus
	workingCount uint
	workers      []*Worker
}

func NewScheduler() (*Scheduler, error) {
	scheduler := &Scheduler{}
	return scheduler, nil
}

func (s *Scheduler) AddWorker(w *Worker) {
	s.workers = append(s.workers, w)
}

func (s *Scheduler) Run(gitClient GitlabWrapper, slack SlackWrapper, config Config, cache *localCache) {
	s.requests = make(chan MergeRequest, 10000)
	defer close(s.requests)

	s.responses = make(chan MRResponse, 100)
	defer close(s.responses)

	s.status = make(chan WorkerStatus, 100)
	defer close(s.status)

	for i, worker := range s.workers {
		log.Debugf("schedule worker: starting : %d.", i)
		go worker.Run(s.requests, s.responses, s.status, gitClient, slack, config, cache)
	}

	s.messagePump()
}

func (s *Scheduler) messagePump() {
	for {
		select {
		case response := <-s.responses:
			log.Debugf("schedule worker: mr processed: %s.", response.status)
			s.handleResponse(response)
		case status := <-s.status:
			s.adjustStatus(status)
			log.Debugf("schedule worker: working: %d. request queue size %d.", s.workingCount, len(s.requests))
		}
	}
}

func (s *Scheduler) handleResponse(response MRResponse) {
}

func (s *Scheduler) adjustStatus(status WorkerStatus) {
	switch status {
	case WorkerWorking:
		s.workingCount++
		promWorkersWorking.Inc()
	case WorkerWaiting:
		s.workingCount--
		promWorkersWorking.Dec()
	default:
		log.Warnf("scheduler worker: unknown status: %d.", status)
	}
}
