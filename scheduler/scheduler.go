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
	Filter      string
	WorkloadId  string
	UploadUrl   string
	DownloadUrl string
	ImageId     string
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
	r, err := c.SayHello(ctx, &pb.HelloRequest{WorkloadId: job.WorkloadId, Filter: job.Filter, UploadUrl: job.UploadUrl, DownloadUrl: job.DownloadUrl, ImageId: job.ImageId})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Scheduler: RPC respose from %s : %s", job.WorkloadId, r.GetImageId())

}

func Start(jobs chan Job) error {
	for {
		job := <-jobs
		// ## if to schedule
		schedule(job)
	}
	return nil
}
