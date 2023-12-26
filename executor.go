package main

type HandlerFunction func(*Job)error

type ExecutionResult struct {Job *Job; Error error}

type Executor struct {
	Handlers map[string]HandlerFunction
	ExecutionChannel chan *Job
	ResultsChannel chan ExecutionResult
}

func NewExecutor(bufsize int) *Executor {
	return &Executor{
		make(map[string]HandlerFunction, bufsize),
		make(chan *Job),
		make(chan ExecutionResult),
	}
}

// Public
func (e *Executor) RegisterHandler(name string, fn HandlerFunction){
	e.Handlers[name] = fn
}

func (e *Executor) Submit(j *Job) {
	e.ExecutionChannel <- j
}

func (e *Executor) Start() {
	// Start job executions
	for job := range e.ExecutionChannel {
		go e.execute(job)
	}
}

// Private
func (e *Executor) execute(j *Job) { 
	err := e.Handlers[j.Handler](j)
	e.ResultsChannel <- ExecutionResult{j, err}
	return
}
