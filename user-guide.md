# User Guide

## Installation:

1. Install Golang following the next instrucctions:
   [Install Golang](https://golang.org/doc/install)
2. Install all dependencies runing the following command:

```
go get ./...
```

## Building and Running Server:

- **To run the server in dev mode run the following command:**

```
go run main.go
```

- **To build the server for production use the following command:**

```
go build main.go
```

## API Usage:

- **To login make an http get request to the next url passing Username and Password in the headers, if the operation is successful you will get the next output format**

```
REQ: <Server host>:8080/login
RES:
{
    "token": <Token>
}
```

- **To logout make an http get request to the next url passing the bearer token in the headers, if the operation is successful you will get the next output format**

```
REQ: "Authorization: Bearer <ACCESS_TOKEN>" <Server host>:8080/logout
RES:
{
	"message": "Bye <Username>, your token has been revoked"
}
```

- **To get the status make an http get request to the next url passing the bearer token in the headers, if the operation is successful you will get the next output format**

```
REQ: "Authorization: Bearer <ACCESS_TOKEN>" <Server host>:8080/status
RES:
{
	"message": "Hi <Username>, the DPIP System is Up and Running"
	"time": <Time Stamp in the server>
}
```

- **To upload an image make an http get request to the next url passing the bearer token in the headers and the file in the request body, if the operation is succesfull you will get the next output fomart**

```
REQ: "Authorization: Bearer <ACCESS_TOKEN>" <Server host>:8080/upload
RES:
{
	"message": "An image has been successfully uploaded",
	"filename": <File Name>,
	"size": <File size>
}
```
- **To request the system status about the workers, pass the bearer token in the headers and make a request to the next url and the next output format will appear**

```
REQ: "Authorization: Bearer <ACCESS_TOKEN>" <Server host>:8080/status/workers
RES:
{
    [
        {
            "name": "cubomx",
            "tag": "gpu,superCPU,largeMemory",
            "ip": "192.168.15.15",
            "port": "50051",
            "lastUpdate": "2020-05-04 00:14:11-05:00",
            "powerStatus": "off",
            "usage": "25"
        },
        {
            "name": "santt",
            "tag": "gpu,superCPU,largeMemory",
            "ip": "192.168.15.15",
            "port": "50051",
            "lastUpdate": "2020-05-04 00:12:39-05:00",
            "powerStatus": "off",
            "usage": "34"
        }
    ]
}
```

- **To make a test of the system from the API to one worker, pass the bearer token in the headers and make a request to the next url and should expect the next output format**

```
REQ: "Authorization: Bearer <ACCESS_TOKEN>" <Server host>:8080/workloads/test
RES:
{
	"message": "hola"
}
```