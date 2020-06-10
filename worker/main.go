package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"nanomsg.org/go/mangos/v2"
	"nanomsg.org/go/mangos/v2/protocol/respondent"

	pb "github.com/Santt99/cool-image-processor/proto"
	// register transports
	"google.golang.org/grpc"
	_ "nanomsg.org/go/mangos/v2/transport/all"
	mem "github.com/shirou/gopsutil/mem"
	cpu "github.com/shirou/gopsutil/cpu"
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
	downloadResponse, downloadErr := http.PostForm(in.GetDownloadUrl(), url.Values{"image-id": {in.GetImageId()}})
	if downloadErr != nil {
		fmt.Print(downloadErr)
	}
	defer downloadResponse.Body.Close()
	createFile(in.GetImageId(), downloadResponse.Body)
	app := "./process"
	output_fn := "out_" + in.GetImageId()
	cmd := exec.Command(app, in.GetFilter(), in.GetImageId(), output_fn)
	_, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
	}
	curl := "curl"
	cmd = exec.Command(curl, "-H", "'content-type: multipart/form-data'", "-F", "workload-id="+in.GetWorkloadId(), "-F", "data=@"+output_fn, in.GetUploadUrl())
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(out))
	return &pb.HelloReply{WorkloadId: in.GetWorkloadId(), ImageId: in.GetImageId(), Filter: in.GetFilter(), DownloadUrl: in.GetDownloadUrl(), UploadUrl: in.GetUploadUrl()}, nil
}
func createFile(fileName string, file io.ReadCloser) error {
	out, err := os.Create("./" + fileName)
	if err != nil {
		fmt.Println(err)
	}
	defer out.Close()
	_, erro := io.Copy(out, file)
	if erro != nil {
		fmt.Println(err)
	}
	return nil
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

func getUsage()(string, string, error){
	v, err := mem.VirtualMemory()

	memUsage := v.UsedPercent
	if (err != nil){
		return "", "", err
	}
	c, err1 := cpu.Times(false)

	if err1 != nil {
		fmt.Printf("%v",err1)
		return "", "", err1
	}
	
	cpuUsage := 0.0
	for _, item := range c {
		cpuUsage = ((item.User + item.System ) / item.Total()) * 100
	}
	return strconv.FormatFloat(cpuUsage, 'f', -1, 32), strconv.FormatFloat(memUsage, 'f', -1, 32), nil
	
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

		cpuUsage, memUsage, err := getUsage()

		if err != nil {
			panic(err)
		}
		workerMetadata := workerName + "@" + tags + "@" + getIP() + "@" 
		workerMetadata += strconv.Itoa(port) + "@" + tf + "@" + cpuUsage + "@" + memUsage
		if err = sock.Send([]byte(workerMetadata)); err != nil {
			die("Cannot send: %s", err.Error())
		}
	}
}
