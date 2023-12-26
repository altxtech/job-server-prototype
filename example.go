package main

import(
	"log"
	"time"
	"math/rand"
	"fmt"
)

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




func main() {
	
	database := NewInMemoryDatabase()
	server := NewJobServer(database)

	server.RegisterHandler("dummy", DummyJob)

	server.Run()
}
