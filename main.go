package main

import (
	"log"
	"github.com/Santt99/cool-image-processor/controller"
	"github.com/Santt99/cool-image-processor/scheduler"
	api "github.com/Santt99/cool-image-processor/api"
)

func main() {
	log.Println("Welcome to the Distributed and Parallel Image Processing System")

	// Start Controller
	go controller.Start()

	// Start Scheduler
	jobs := make(chan scheduler.Job)
	jobsByCodeName := make(chan int, 10)
	go scheduler.Start(jobs)

	// Send sample jobs
	go api.Run(jobsByCodeName)
	for {
		select {
			case data := <-jobsByCodeName:
				if data == 1{
					currentJob := scheduler.Job{Work: "SayHello", RPCName: "Paco"}
					jobs <- currentJob
				}
			default:
		}
	}
	
	
	
	
	// API

	
	// Here's where your API setup will be
}
