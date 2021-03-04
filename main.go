package main

import (
	"log"

	api "github.com/Santt99/cool-image-processor/api"
	"github.com/Santt99/cool-image-processor/controller"
	"github.com/Santt99/cool-image-processor/scheduler"
)

func main() {
	log.Println("Welcome to the Distributed and Parallel Image Processing System")

	// Start Controller
	go controller.Start()

	// Start Scheduler
	jobs := make(chan scheduler.Job)
	jobsByCodeName := make(chan api.FilterJob, 10)
	go scheduler.Start(jobs)

	// Send sample jobs
	go api.Run(jobsByCodeName)
	for {
		select {
		case data := <-jobsByCodeName:

			currentJob := scheduler.Job{Filter: data.Filter, WorkloadId: data.WorkloadID, UploadUrl: "http://localhost:8080/upload", DownloadUrl: "http://localhost:8080/download", ImageId: data.ImageID}
			jobs <- currentJob

		default:
		}
	}

	// API

	// Here's where your API setup will be
}
