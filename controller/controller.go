package controller

import (
	"fmt"
	"os"
	"time"

	// "encoding/json"
	// "bytes"
	"log"
	"strings"

	bolt "go.etcd.io/bbolt"
	"nanomsg.org/go/mangos/v2"
	"nanomsg.org/go/mangos/v2/protocol/surveyor"

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

func Start() {
	var sock mangos.Socket
	var err error
	var msg []byte
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	_, err = tx.CreateBucketIfNotExists([]byte("Workers"))
	if err != nil {
		log.Fatal(err)
	}

	if sock, err = surveyor.NewSocket(); err != nil {
		die("can't get new surveyor socket: %s", err)
	}
	if err = sock.Listen(controllerAddress); err != nil {
		die("can't listen on surveyor socket: %s", err.Error())
	}
	err = sock.SetOption(mangos.OptionSurveyTime, time.Second/2)
	if err != nil {
		die("SetOption(): %s", err.Error())
	}
	for {
		time.Sleep(time.Second)
		fmt.Println("SERVER: SENDING DATE SURVEY REQUEST")
		if err = sock.Send([]byte("DATE")); err != nil {
			die("Failed sending survey: %s", err.Error())
		}
		for {
			if msg, err = sock.Recv(); err != nil {
				break
			}
			workerMetadata := strings.Split(string(msg), "@")
			workerName := workerMetadata[0]
			lastUpdate := workerMetadata[4]
			insertWorkerToDB(Worker{workerName, workerMetadata[1], workerMetadata[2], workerMetadata[3], lastUpdate, "On"})
		}
		updateWorkersPowerStatus()
		fmt.Println("SERVER: SURVEY OVER")
	}
}

type Worker struct {
	Name        string `json:"name"`
	Tag         string `json:"tag"`
	IP          string `json:"ip"`
	Port        string `json:"port"`
	LastUpdate  string `json:"lastUpdate"`
	PowerStatus string `json:"powerStatus"`
}

func insertWorkerToDB(worker Worker) {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	workers := tx.Bucket([]byte("Workers"))

	node, err := workers.CreateBucketIfNotExists([]byte(worker.Name))
	if err != nil {
		log.Fatal(err)
	}
	// Setup the users bucket.
	nodeMetadata, err := node.CreateBucketIfNotExists([]byte("metadata"))
	if err != nil {
		log.Fatal(err)
	}

	err = nodeMetadata.Put([]byte("name"), []byte(worker.Name))
	if err != nil {
		log.Fatal(err)
	}

	err = nodeMetadata.Put([]byte("tag"), []byte(worker.Tag))
	if err != nil {
		log.Fatal(err)
	}

	err = nodeMetadata.Put([]byte("ip"), []byte(worker.IP))
	if err != nil {
		log.Fatal(err)
	}

	err = nodeMetadata.Put([]byte("port"), []byte(worker.Port))
	if err != nil {
		log.Fatal(err)
	}
	err = nodeMetadata.Put([]byte("timestamp"), []byte(worker.LastUpdate))
	if err != nil {
		log.Fatal(err)
	}
	err = nodeMetadata.Put([]byte("powerStatus"), []byte("On"))
	if err != nil {
		log.Fatal(err)
	}
	// Commit the transaction and check for error.
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func GetWorkers() []Worker {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	workers := tx.Bucket([]byte("Workers"))
	workersCursor := workers.Cursor()
	workersLength := workersCursor.Bucket().Stats().InlineBucketN
	if workersLength <= 0 {
		return nil
	}
	workersMetadataArray := make([]Worker, workersLength)
	for key, _ := workersCursor.First(); key != nil; key, _ = workersCursor.Next() {

		node := workers.Bucket([]byte(key))
		e := node.Bucket([]byte("metadata"))
		c := e.Cursor()
		nodeName, nodeIp, nodePort, nodePowerStatus, nodeTimestamp, nodeTags := "", "", "", "", "", ""

		index := 0
		for key, value := c.First(); key != nil; key, value = c.Next() {
			switch key := string(key); key {
			case "name":
				nodeName = string(value)
			case "ip":
				nodeIp = string(value)
			case "port":
				nodePort = string(value)
			case "powerStatus":
				nodePowerStatus = string(value)
			case "tag":
				nodeTags = string(value)
			default:
				nodeTimestamp = string(value)
			}
		}

		workersMetadataArray[index] = Worker{nodeName, nodeTags, nodeIp, nodePort, nodeTimestamp, nodePowerStatus}
		index = index + 1
	}
	return workersMetadataArray
}

func updateWorkersPowerStatus() {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	workers := tx.Bucket([]byte("Workers"))
	workersCursor := workers.Cursor()

	for key, _ := workersCursor.First(); key != nil; key, _ = workersCursor.Next() {
		node := workers.Bucket([]byte(key))
		e := node.Bucket([]byte("metadata"))
		c := e.Cursor()

		for key, value := c.First(); key != nil; key, value = c.Next() {
			if string(key) == "timestamp" {
				layout := "2006-01-02 15:04:05-07:00"
				lastUpdate, err := time.Parse(layout, string(value))
				if err != nil {
					log.Fatal(err)
				}
				t := time.Now()
				diff := t.Sub(lastUpdate)
				if diff > 5000 {
					e.Put([]byte("powerStatus"), []byte("off"))
				}
			}
		}
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
