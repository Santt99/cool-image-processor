package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"nanomsg.org/go/mangos/v2"
	"nanomsg.org/go/mangos/v2/protocol/respondent"

	pb "github.com/Santt99/cool-image-processor/proto"
	// register transports
	"google.golang.org/grpc"
	_ "nanomsg.org/go/mangos/v2/transport/all"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

var (
	defaultRPCPort = 50051
)

var (
	controllerAddress = ""
	workerName        = ""
	tags              = ""
)

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func date() string {
	return time.Now().Format(time.ANSIC)
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("RPC: Received: %v", in.GetWorkloadId())
	return &pb.HelloReply{Message: "Hello " + in.GetWorkloadId() + " " + in.GetImageId() + " " + in.GetFilter() + " " + in.GetUploadUrl() + " " + in.GetDownloadUrl()}, nil
}

func init() {
	flag.StringVar(&controllerAddress, "controller", "tcp://localhost:40899", "Controller address")
	flag.StringVar(&workerName, "worker-name", "hard-worker", "Worker Name")
	flag.StringVar(&tags, "tags", "gpu,superCPU,largeMemory", "Comma-separated worker tags")
}

func main() {
	flag.Parse()
	// Subscribe to Controller
	go joinCluster()
	// Setup Worker RPC Server
	rpcPort := getAvailablePort()
	log.Printf("Starting RPC Service on localhost:%v", rpcPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", rpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func getAvailablePort() int {
	port := defaultRPCPort
	for {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
		if err != nil {
			port = port + 1
			continue
		}
		ln.Close()
		break
	}
	return port
}

func getIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		localAddr := "unknown"
		return localAddr
	}

	defer conn.Close()
	localAddr := strings.Split(conn.LocalAddr().(*net.UDPAddr).String(), ":")[0]
	return localAddr
}

func joinCluster() {
	var sock mangos.Socket
	var err error
	var msg []byte
	var name = "1"
	if sock, err = respondent.NewSocket(); err != nil {
		die("can't get new respondent socket: %s", err.Error())
	}
	if err = sock.Dial(controllerAddress); err != nil {
		die("can't dial on respondent socket: %s", err.Error())
	}
	for {
		if msg, err = sock.Recv(); err != nil {
			die("Cannot recv: %s", err.Error())
		}
		fmt.Printf("CLIENT(%s): RECEIVED \"%s\" SURVEY REQUEST\n",
			name, string(msg))
		port := getAvailablePort()
		fmt.Printf("CLIENT(%s): SENDING DATE SURVEY RESPONSE\n", name)
		t := time.Now()
		tf := t.Format("2006-01-02 15:04:05-07:00")
		usage, err := rand.Int(rand.Reader, big.NewInt(100))

		if err != nil {
			panic(err)
		}
		workerMetadata := workerName + "@" + tags + "@" + getIP() + "@" + strconv.Itoa(port) + "@" + tf + "@" + usage.String()
		if err = sock.Send([]byte(workerMetadata)); err != nil {
			die("Cannot send: %s", err.Error())
		}
	}
}
