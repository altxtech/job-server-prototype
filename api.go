
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
)

type CreateJobRequest struct {
	// ID and status should be auto generated
	Handler string `json:"handler"`
	MaxAttempts int `json:"max_attempts"`
}


// /jobs
func (s *JobServer) CreateJob() gin.HandlerFunc {
	fn := func (c *gin.Context){

		var req CreateJobRequest
		err := c.ShouldBindJSON(&req)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err)
			return
		}

		// Check if it is a valid handler
		_, err = s.Database.GetHandlerByName(req.Handler)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid handler '%s'", req.Handler)})
			return
		}

		// Define the id (Auto increment)
		newJob := NewJob(req.Handler, req.MaxAttempts)
		newJob, err = s.Database.CreateJob(newJob)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error adding job to database: %v", err)})
			return
		}

		// Submit for execution
		s.Submit(newJob)

		c.IndentedJSON(http.StatusCreated, newJob)
		return
	}

	return fn
}

func (s *JobServer) ListJobs() gin.HandlerFunc {
	fn := func(c *gin.Context){
		jobs, err := s.Database.GetAllJobs() 
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error getting jobs from database: %v", err)})
			return
		}
		c.IndentedJSON(http.StatusOK, jobs)
	}
	return fn
}

func (s *JobServer) GetJob() gin.HandlerFunc {
	fn := func (c *gin.Context){
		jobId, err := strconv.ParseInt(c.Param("jobId"), 10, 32)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err)
		}
		job, err := s.Database.GetJobById(jobId) 
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Job with id %d not found", jobId)})
			return
		}
		c.IndentedJSON(http.StatusFound, job)
		return
	}
	return fn
}

func (s *JobServer) UpdateJob() gin.HandlerFunc {
	/*
		This method should also allow for changing the status of the Job -> Use case: job cancellation
	*/
	fn := func (c *gin.Context){
		var req CreateJobRequest // Can also be used for patching
		err := c.ShouldBindJSON(&req)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err)
			return
		}

		jobId, err := strconv.ParseInt(c.Param("jobId"), 10, 32)
		updatedJob := NewJob(req.Handler, req.MaxAttempts)
		updatedJob.SetID(jobId)
		updatedJob, err = s.Database.UpdateJob(updatedJob)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error writing updates to database: %v", err)})
			return
		}
		
		c.IndentedJSON(http.StatusOK, updatedJob)
		return
	}
	return fn
}


// /handlers
/* 
	Job handlers are registered before the API server starts.
	The user can list available handlers, but cannot create, edit or delete them
*/
func (s *JobServer) ListHandlers() gin.HandlerFunc {
	fn := func ( c *gin.Context ) {
		handlers, err := s.Database.GetAllHandlers()
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error retrieving handlers: %v", err)})
			return
		}
		c.IndentedJSON(http.StatusOK, handlers)
		return
	}
	return fn
}

func (s *JobServer) GetHandler() gin.HandlerFunc {
	fn := func ( c *gin.Context ) {
		handlerId, err := strconv.ParseInt(c.Param("handlerId"), 10, 32)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message":fmt.Sprintf("Invalid id %s", c.Param("handlerId"))})
		}
		handler, err := s.Database.GetHandlerById(handlerId)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Handler with id %d not found", handlerId)})
			return
		}
		c.IndentedJSON(http.StatusFound, handler)
		return
	}
	return fn
}

func (s *JobServer) CreateRouter() *gin.Engine {

	// Register API handlers
	router := gin.Default()

	// Jobs
	router.POST("/api/jobs", s.CreateJob())
	router.GET("/api/jobs", s.ListJobs())
	router.GET("/api/jobs/:jobId", s.GetJob())
	router.PATCH("/api/jobs/:jobId", s.UpdateJob())

	// Job Handlers
	router.GET("/api/handlers", s.ListHandlers())
	router.GET("/api/handlers/:handlerId", s.GetHandler())

	return router
}
