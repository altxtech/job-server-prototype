package main


type Database interface {
	// Handlers
	CreateHandler(*Handler) (*Handler, error)
	GetHandlerByName(string) (*Handler, error)
	GetHandlerById(int64) (*Handler, error)
	GetAllHandlers() ([]Handler, error)

	// Jobs
	CreateJob(*Job) (*Job, error)
	GetAllJobs() ([]Job, error)
	GetJobById(int64) (*Job, error)
	UpdateJob(*Job) (*Job, error)
}
