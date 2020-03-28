package main

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var secrets = gin.H{
	"foo":    gin.H{"email": "foo@bar.com", "phone": "123433", "lastToken": ""},
	"austin": gin.H{"email": "austin@example.com", "phone": "666", "lastToken": ""},
	"lena":   gin.H{"email": "lena@guapa.com", "phone": "523443", "lastToken": ""},
}

var tokens = make(map[string]string)

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

type Token struct {
	Token string `json:token`
}

type Err struct {
	code    int
	message string
}

func (e *Err) Error() string {
	return fmt.Sprintf("%d-%s", e.code, e.message)
}

func main() {
	r := gin.Default()

	r.Use()
	// Group using gin.BasicAuth() middleware
	// gin.Accounts is a shortcut for map[string]string
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":    "bar",
		"austin": "1234",
		"lena":   "hello2",
		"manu":   "4321",
	}))

	authorized.GET("/login", login)
	r.GET("/user", getUser)
	r.GET("/logout", logout)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
func logout(c *gin.Context) {
	username, err := auth(c)
	if err != nil {
		errorCode := getErrorCode(err)
		c.AbortWithStatus(errorCode)
		return
	}
	tokens[username] = ""
	c.AbortWithStatus(200)
	return
}
func getErrorCode(err error) int {
	errorParts := strings.Split(err.Error(), "-")
	errorCode, err := strconv.Atoi(errorParts[0])
	if err != nil {
		return 500
	}
	return errorCode
}
func getUser(c *gin.Context) {
	username, err := auth(c)
	if err != nil {
		errorCode := getErrorCode(err)
		c.AbortWithStatus(errorCode)
		return
	}
	c.JSON(200, gin.H{"username": username})
}

func auth(c *gin.Context) (string, error) {
	var token Token
	if err := c.ShouldBindJSON(&token); err != nil {
		return "", &Err{400, "Bad Request"}
	}

	// Get the JWT string from the cookie
	tknStr := token.Token

	// Initialize a new instance of `Claims`
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		fmt.Print(reflect.TypeOf(err))
		if tkn == nil {
			return "", &Err{400, "Bad Request"}
		}
		return "", &Err{401, "Token not valid"}
	}

	if tk, ok := tokens[claims.Username]; ok {
		if tk != tknStr {
			return "", &Err{401, "Token not valid"}
		}
	} else {
		return "", &Err{400, "Bad Request"}
	}

	return claims.Username, nil
}

func login(c *gin.Context) {

	// get user, it was set by the BasicAuth middleware
	user := c.MustGet(gin.AuthUserKey).(string)

	// /admin/secrets endpoint
	// hit "localhost:8080/admin/secrets
	expirationTime := time.Now().Add(5 * time.Minute)
	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		Username: user,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.AbortWithStatus(500)
	}
	tokens[user] = tokenString
	if _, ok := secrets[user]; ok {
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	} else {
		c.AbortWithStatus(401)
	}

}
