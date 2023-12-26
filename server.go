package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

/*
	A Job Server is formed by the following:

	- An API, which ingests incoming http requests,

	- A database interface

	- And executor, which runs the jobs

	API      DATABASE       EXECUTOR
	|---------->|-------------->|
	|           |<--------------|
*/
// Handlers
type Handler struct {
	ID int64 `json:"id"`
	Name string `json:"name"`
}

type HandlerFunction func(*Job)error
// Jobs
/*
	Job statuses:
	READY - ready to be executed
	RUNNING - currently being executed
	TIMEOUT - jobs has failed and waiting to be retried
	SUCCEDED - job finished successfully
	FAILED - exceeded maximum retries

*/
type Job struct {
	ID int64 `json:"id"`
	Handler string `json:"handler"`
	Status string `json:"status"`
	Attempts int `json:"attempts"`
	MaxAttempts int `json:"max_attempts"`
	Error string `json:"error"`
}

func NewJob(handler string, maxAttempts int) *Job {
	return &Job{
		ID: 0,
		Handler: handler,
		Status: "READY",
		Attempts: 0,
		MaxAttempts: maxAttempts, 
		Error: "",
	}
}

func (j *Job) SetID(id int64) {
	j.ID = id
}

type JobServer struct {
	Router *gin.Engine
	Database Database
	Handlers map[string]HandlerFunction
	ExecutionChannel chan *Job
}


func NewJobServer(db Database) *JobServer{
	server := &JobServer{
		Database: db,
		Handlers: make(map[string]HandlerFunction),
		ExecutionChannel: make(chan *Job, 10),
	}
	server.Router = server.CreateRouter()

	return server
}


// At server startup

func (s *JobServer) RegisterHandler(name string, fn HandlerFunction){
	
	// Add to the executor 
	s.Handlers[name] = fn
	
	// Add to the Database
	newHandler := &Handler{
		ID: 0, // Will be attributed when inserted into the database
		Name: name,
	}
	newHandler, err := s.Database.CreateHandler(newHandler)
	if err != nil {
		log.Fatal(err)
	}

	// Handler should have been attributed a non zero ID
	if newHandler.ID == 0 {
		log.Fatal("Handler incorrectly inserted into database. Id is 0")
	} else {
		log.Printf("Handler '%s' successfully created with Id %d", newHandler.Name, newHandler.ID)
	}
}

func (s *JobServer) Submit(j *Job) {
	s.ExecutionChannel <- j
}

func (s *JobServer) startJobExecution() {
	// Start job executions
	for job := range s.ExecutionChannel {
		go s.execute(job)
	}
}


func (s *JobServer) execute(j *Job) {

	// If it was successfull, mark the job as completed
	j.Status = "RUNNING"
	s.Database.UpdateJob(j)
	jobErr := s.Handlers[j.Handler](j)
	if jobErr == nil {
		j.Status = "SUCCEDED"
		_, err := s.Database.UpdateJob(j)
		if err != nil {
			log.Printf("Error updating Job %d in database %v\n", j.ID, err)
		}
		return
	}

	// If it failed, we need to assess if we need to retry it
	if j.Attempts >= j.MaxAttempts {
		j.Status = "FAILED"
		j.Error = fmt.Sprint(jobErr)
		s.Database.UpdateJob(j)
	}

	j.Status = "TIMEOUT"
	s.Database.UpdateJob(j)
	time.Sleep(time.Duration(100) * time.Millisecond)
	
	j.Status = "READY" 
	j.Attempts += 1 
	s.Database.UpdateJob(j)
	s.Submit(j)
}


func (s *JobServer) Run() {
	go s.startJobExecution()
	go s.Router.Run()
}

