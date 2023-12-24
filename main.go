package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	// Packages to simulate usage
	"time"
	"math/rand"
	"log"
)

// Handlers
type Handler struct {
	ID int `json:"id"`
	Name string `json:"name"`
}

// Jobs
type Job struct {
	ID int `json:"id"`
	Handler string `json:"handler"`
	Status string `json:"status"`
	Attempts int `json:"attempts"`
	MaxAttempts int `json:"max_attempts"`
	Error string `json:"error"`
}

var execChannel chan *Job = make(chan *Job, 10)

/*
	Job statuses:
	READY - ready to be executed
	RUNNING - currently being executed
	TIMEOUT - jobs has failed and waiting to be retried
	SUCCEDED - job finished successfully
	FAILED - exceeded maximum retries

*/


type CreateJobRequest struct {
	// ID and status should be auto generated
	Handler string `json:"handler"`
}

/*
	DUMMY DATABASE
*/
var jobs = []*Job{}
var handlers = []*Handler{}



/*
	API
*/
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


/*
	JOB EXECUTION MANAGEMENT
*/
var handlersMap = map[string]func(*Job)error{}

func RegisterJobHandler(name string, fn func(*Job)error){
	newHandler := &Handler{
		ID: len(handlers) + 1,
		Name: name,
	}
	handlersMap[name] = fn
	handlers = append(handlers, newHandler)
}

func ExecuteJob(j *Job) {
	j.Status = "RUNNING"
	j.Attempts += 1
	err := handlersMap[j.Handler](j)
	if err != nil {
		j.Error = fmt.Sprint(err)
		// Check if the job should be retried 
		if j.Attempts >= j.MaxAttempts {
			j.Status = "FAILED"
			return
		}

		// Timeout
		j.Status = "TIMEOUT"
		time.Sleep(time.Duration(100) * time.Millisecond)
		j.Status = "READY"
		execChannel <- j
		return
	}

	j.Status = "SUCCEDED"
	return
}

func StartJobExecution() {
	// Start job executions
	for job := range execChannel {
		go ExecuteJob(job)
	}
}


/*
	EXAMPLE USAGE
*/
func DummyJob (j *Job) error {
	// The job takes some time to finish...
	log.Printf("Staring jobs %d\n", j.ID)
	time.Sleep(time.Duration(10) * time.Second)

	// And has a chance to fail
	if rand.Intn(10) == 0 {
		log.Printf("Job %d failed", j.ID)
		return fmt.Errorf("Job %d failed", j.ID)
	}
	log.Printf("Job %d succeded", j.ID)

	return nil
}

func ExampleUsage() {
	
	// This would be similar to the user's main() function
	
	// db := jobserver.SomeDBImplementation() - For example, postgressql or firestore, or whatever is the user's preference for a backend
	// server := jobserver.NewServer()

	RegisterJobHandler("dummy", DummyJob) // server
	
	// server.Run("localhost:8080")
}



func main() {
	router := gin.Default()

	// Jobs
	router.POST("/api/jobs", CreateJob)
	router.GET("/api/jobs", ListJobs)
	router.GET("/api/jobs/:jobId", GetJob)
	router.PATCH("/api/jobs/:jobId", UpdateJob)

	// Job Handlers
	router.GET("/api/handlers", ListHandlers)
	router.GET("/api/handlers/:handlerId", GetHandler)

	ExampleUsage()
	go StartJobExecution()

	router.Run("localhost:8080")
}
