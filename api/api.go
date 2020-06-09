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

	"path/filepath"

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
var jobsQueue = make(chan FilterJob, 10)

// Create the JWT key used to create the signature
var jwtKey = []byte("my_secret_key")

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

func Run(jobs chan FilterJob) {
	r := gin.Default()
	r.LoadHTMLGlob(filepath.Join(os.Getenv("GOPATH"), "/src/github.com/Santt99/cool-image-processor/api/templates/*"))

	jobsQueue = jobs
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":    "bar",
		"austin": "1234",
		"lena":   "hello2",
		"manu":   "4321",
	}))

	authorized.POST("/login", login)
	r.StaticFS("/results", gin.Dir("./results", true))
	r.GET("/results", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"tree": getResultsTree(),
		})
	})
	r.GET("/status", getWorkersStatus)
	r.GET("/status/:worker", getWorkerStatus)
	r.GET("/logout", logout)
	r.POST("/upload", upload)
	// r.GET("/download", download)
	r.POST("/workloads/filter", filter)
	r.Run(":8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func getResultsTree() []map[string]string {
	var files []map[string]string

	root := "./results/"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if path != root {
			file := make(map[string]string)
			file["name"] = info.Name()
			file["path"] = path
			file["isDir"] = strconv.FormatBool(info.IsDir())
			files = append(files, file)
		}
		fmt.Println(info)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

func getWorkerStatus(c *gin.Context) {
	workerName := c.Param("worker")
	worker := controller.GetWorker(workerName)
	if (worker == controller.Worker{}) {
		c.JSON(http.StatusOK, gin.H{"message": "No workers registered"})
	}
	c.JSON(http.StatusOK, worker)
}

type FilterJob struct {
	WorkloadID string `json:"workload-id"`
	Filter     string `json:"filter"`
	ImageID    string
}

func filter(c *gin.Context) {
	_, err := auth(c)
	if err != nil {
		errorCode := getErrorCode(err)
		c.AbortWithStatus(errorCode)
		return
	}
	workloadId := c.Request.FormValue("workload-id")

	filter := c.Request.FormValue("filter")

	file, header, err := c.Request.FormFile("data")
	if err != nil {
		returnError(err, c)
		return
	}
	err = createFile("./uploads/"+header.Filename, file)
	if err != nil {
		returnError(err, c)
		return
	}
	// Pass filterJob.WorkloadID
	filterJob := FilterJob{workloadId, filter, header.Filename}
	jobsQueue <- filterJob
	resultsURL := fmt.Sprint("http://localhost:8080", "/results/", filterJob.WorkloadID)
	c.JSON(http.StatusOK, gin.H{"Workload ID": filterJob.WorkloadID, "Filter": filterJob.Filter, "Status": "Scheduling", "Results": resultsURL, "Filename": header.Filename})
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
	//Todo: Add worload_id subdirectort
	err = createFile("./results/"+header.Filename, file)
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
