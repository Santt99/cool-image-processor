package scheduler

import (
	"context"
	"log"
	"time"

	pb "github.com/Santt99/cool-image-processor/proto"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

type Job struct {
	Work string
	RPCName string
}

func schedule(job Job) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if(job.Work == "SayHello"){
		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: job.RPCName})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Scheduler: RPC respose from %s : %s", job.Work, r.GetMessage())
	}
	
	
}

func Start(jobs chan Job) error {
	for {
		job := <-jobs
		// ## if to schedule
		schedule(job)
	}
	return nil
}
