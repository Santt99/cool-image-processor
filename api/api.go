package api

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Santt99/cool-image-processor/controller"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var secrets = gin.H{
	"foo":    gin.H{"email": "foo@bar.com", "phone": "123433", "lastToken": ""},
	"austin": gin.H{"email": "austin@example.com", "phone": "666", "lastToken": ""},
	"lena":   gin.H{"email": "lena@guapa.com", "phone": "523443", "lastToken": ""},
}

var tokens = make(map[string]string)
var jobsQueue = make(chan int, 10)

// Create the JWT key used to create the signature
var jwtKey = []byte("my_secret_key")

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type Err struct {
	code    int
	message string
}

func (e *Err) Error() string {
	return fmt.Sprintf("%d-%s", e.code, e.message)
}

func Run(jobs chan int) {
	r := gin.Default()
	r.Use()
	jobsQueue = jobs
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":    "bar",
		"austin": "1234",
		"lena":   "hello2",
		"manu":   "4321",
	}))

	authorized.GET("/login", login)
	r.GET("/status", getWorkersStatus)
	r.GET("/status/:worker", getWorkerStatus)
	r.GET("/logout", logout)
	r.GET("/upload", upload)
	r.GET("/workloads/test", hello)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
func getWorkerStatus(c *gin.Context) {
	workerName := c.Param("worker")
	worker := controller.GetWorker(workerName)
	if (worker == controller.Worker{}) {
		c.JSON(http.StatusOK, gin.H{"message": "No workers registered"})
	}
	c.JSON(http.StatusOK, worker)
}
func hello(c *gin.Context) {
	_, err := auth(c)
	if err != nil {
		errorCode := getErrorCode(err)
		c.AbortWithStatus(errorCode)
		return
	}
	jobsQueue <- 1
	c.JSON(http.StatusOK, gin.H{"message": "Hola"})
}

func getWorkersStatus(c *gin.Context) {
	_, err := auth(c)
	if err != nil {
		errorCode := getErrorCode(err)
		c.AbortWithStatus(errorCode)
		return
	}
	workers := controller.GetWorkers()
	if workers == nil {
		c.JSON(http.StatusOK, gin.H{"message": "No workers registered"})
	}
	c.JSON(http.StatusOK, workers)
}
func logout(c *gin.Context) {
	username, err := auth(c)
	if err != nil {
		errorCode := getErrorCode(err)
		c.AbortWithStatus(errorCode)
		return
	}
	tokens[username] = ""
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprint("Bye ", username, ", your token has been revoked")})
	return
}
func getErrorCode(err error) int {
	errorParts := strings.Split(err.Error(), "-")
	errorCode, err := strconv.Atoi(errorParts[0])
	if err != nil {
		return http.StatusInternalServerError
	}
	return errorCode
}
func getStatus(c *gin.Context) {

	username, err := auth(c)
	if err != nil {
		returnError(err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprint("Hi ", username, ", the DPIP System is Up and Running"), "time": time.Now()})
}

func upload(c *gin.Context) {

	_, err := auth(c)
	if err != nil {
		returnError(err, c)
		return
	}

	file, header, err := c.Request.FormFile("data")
	if err != nil {
		returnError(err, c)
		return
	}
	err = createFile(header.Filename, file)
	if err != nil {
		returnError(err, c)
		return
	}

	size := strconv.Itoa(int(header.Size))

	c.JSON(http.StatusOK, gin.H{"status": "SUCCESS", "fileName": header.Filename, "fileSize": size + " bytes"})
}

func returnError(err error, c *gin.Context) {
	errorCode := getErrorCode(err)
	c.AbortWithStatus(errorCode)
}

func createFile(fileName string, file multipart.File) error {
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		return err
	}
	return nil
}

func auth(c *gin.Context) (string, error) {
	bearberToken := c.GetHeader("Authorization")
	token := strings.Split(bearberToken, " ")[1]
	return itExist(token)
}

func itExist(tknStr string) (string, error) {
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if tkn == nil {
			return "", &Err{http.StatusBadRequest, "Bad Request"}
		}
		return "", &Err{http.StatusUnauthorized, "Token not valid"}
	}

	if tk, ok := tokens[claims.Username]; ok {
		if tk != tknStr {
			return "", &Err{http.StatusUnauthorized, "Token not valid"}
		}
	} else {
		return "", &Err{http.StatusUnauthorized, "Token not valid"}
	}
	return claims.Username, nil
}

func login(c *gin.Context) {
	username := c.MustGet(gin.AuthUserKey).(string)

	expirationTime := time.Now().Add(20 * time.Minute) // Here he should modified to give more time to the user

	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.AbortWithStatus(500)
	}
	tokens[username] = tokenString
	if _, ok := secrets[username]; ok {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprint("Hi ", username, ", welcome to the DPIP System"), "token": tokenString})
	} else {
		c.AbortWithStatus(401)
	}
}
