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
	ID int `json:"id"`
	Name string `json:"name"`
}

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
	ID int `json:"id"`
	Handler string `json:"handler"`
	Status string `json:"status"`
	Attempts int `json:"attempts"`
	MaxAttempts int `json:"max_attempts"`
	Error string `json:"error"`
}

type JobServer struct {
	API *gin.Engine
	Database Database
	Executor *Executor
}

func NewJobServer(db Database) *JobServer {
	executor := NewExecutor(10)

	// Register API handlers
	router := gin.Default()

	// Jobs
	router.POST("/api/jobs", CreateJob)
	router.GET("/api/jobs", ListJobs)
	router.GET("/api/jobs/:jobId", GetJob)
	router.PATCH("/api/jobs/:jobId", UpdateJob)

	// Job Handlers
	router.GET("/api/handlers", ListHandlers)
	router.GET("/api/handlers/:handlerId", GetHandler)

	server := &JobServer{
		router,
		db,
		executor,
	}
	return server
}

// At server startup

func (s *JobServer) RegisterHandler(name string, fn HandlerFunction){
	
	// Add to the executor 
	s.Executor.RegisterHandler(name, fn)
	
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

func (s *JobServer) listenForExecutionResults() {
	for result := range s.Executor.ResultsChannel {
		go s.processExecutionResult(result)
	}
}

func (s *JobServer) processExecutionResult(r ExecutionResult) {

	// If it was successfull, mark the job as completed
	if r.Error == nil {
		r.Job.Status = "SUCCEDED"
		err := s.Database.UpdateJob(r.Job)
		if err != nil {
			log.Printf("Error updating Job %d in database %v\n", r.Job.ID, err)
		}

		return
	}

	// If it failed, we need to assess if we need to retry it
	if r.Job.Attempts >= r.Job.MaxAttempts {
		r.Job.Status = "FAILED"
		r.Job.Error = fmt.Sprint(r.Error)
		s.Database.UpdateJob(r.Job)
	}

	r.Job.Status = "TIMEOUT"
	s.Database.UpdateJob(r.Job)
	time.Sleep(time.Duration(100) * time.Millisecond)
	
	r.Job.Status = "RUNNING" 
	r.Job.Attempts += 1 
	s.Database.UpdateJob(r.Job)
	s.Executor.Submit(r.Job)
}


func (s *JobServer) Run() {
	go s.Executor.Start()
	go s.listenForExecutionResults()
	go s.API.Run()
}
