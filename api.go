
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
}

// /jobs
func CreateJob(c *gin.Context){

	var req CreateJobRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}

	// Check if it is a valid handler
	validHandler := false
	for _, handler := range handlers {
		if handler.Name == req.Handler {
			validHandler = true
			break
		}
	}
	if !validHandler {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("\"%s\" is not a valid job handler", req.Handler)})
		return
	}

	// Define the id (Auto increment)
	newJob := &Job{ len(jobs) + 1, req.Handler, "READY", 0, 3, ""}
	jobs = append(jobs, newJob)
	execChannel <- newJob
	c.IndentedJSON(http.StatusCreated, newJob)
	return
}

func ListJobs(c *gin.Context){
	c.IndentedJSON(http.StatusOK, jobs)
}

func GetJob(c *gin.Context){
	jobId, err := strconv.ParseInt(c.Param("jobId"), 10, 32)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
	}
	for _, job := range jobs {
		if job.ID == int(jobId) {
			c.IndentedJSON(http.StatusOK, job)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Job with id %d not found", jobId)})
	return
}

func UpdateJob(c *gin.Context){
	var req CreateJobRequest // Can also be used for patching
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}

	jobId, err := strconv.ParseInt(c.Param("jobId"), 10, 32)
	for idx, job := range jobs {
		if job.ID == int(jobId) {
			jobs[idx].Handler = req.Handler
			c.IndentedJSON(http.StatusOK, jobs[idx])
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Job with id %d not found", jobId)})
	return
}


// /handlers
/* 
	Job handlers are registered before the API server starts.
	The user can list available handlers, but cannot create, edit or delete them
*/

func ListHandlers( c *gin.Context ) {
	c.IndentedJSON(http.StatusOK, handlers)
	return
}

func GetHandler( c *gin.Context ) {
	handlerId, err := strconv.ParseInt(c.Param("handlerId"), 10, 32)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message":fmt.Sprintf("Invalid id %s", c.Param("handlerId"))})
	}
	for _, handler := range handlers {
		if handler.ID == int(handlerId) {
			c.IndentedJSON(http.StatusOK, handler)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Handler with id %d not found", handlerId)})
	return
}
