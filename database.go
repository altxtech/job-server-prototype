package main

import (
	"cloud.google.com/go/firestore"
)


type Database interface {
	// Handlers
	CreateHandler(*Handler) (*Handler, error)

	// Jobs
	CreateJob(*Job) (*Job, error)
	UpdateJob(*Job) (error)
}
