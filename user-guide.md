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

- **To login make an hhtp get request to the next url passing Username and Password in the headers, if the operation is succesfull you will get the next output fomart**

```
REQ: <Server host>:8080/login
RES:
{
    "token": <Token>
}
```

- **To logout make an hhtp get request to the next url passing the beaber token in the headers, if the operation is succesfull you will get the next output fomart**

```
REQ: <Server host>:8080/login
RES:
{
	"message": "Bye <Username>, your token has been revoked"
}
```

- **To get the status make an hhtp get request to the next url passing the beaber token in the headers, if the operation is succesfull you will get the next output fomart**

```
REQ: <Server host>:8080/status
RES:
{
	"message": "Hi <Username>, the DPIP System is Up and Running"
	"time": <Time Stamp in the server>
}
```

- **To upload an image make an hhtp get request to the next url passing the beaber token in the headers and the file in the request body, if the operation is succesfull you will get the next output fomart**

```
REQ: <Server host>:8080/upload
RES:
{
	"message": "An image has been successfully uploaded",
	"filename": <File Name>,
	"size": <File size>
}
```
