package main 


import (
	"fmt"
	"os"
	"time"

	"nanomsg.org/go/mangos/v2"
	"nanomsg.org/go/mangos/v2/protocol/respondent"

	// register transports
	_ "nanomsg.org/go/mangos/v2/transport/all"
)


var controllerAddress = "tcp://localhost:40899"

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func date() string {
	return time.Now().Format(time.ANSIC)
}

func main(){
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

		d := date()
		fmt.Printf("CLIENT(%s): SENDING DATE SURVEY RESPONSE\n", name)
		if err = sock.Send([]byte(d)); err != nil {
			die("Cannot send: %s", err.Error())
		}
	}
}