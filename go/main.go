package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	dbDriver   = "mysql"
	dbUser     = "root"
	dbPassword = "dbpwd"
	dbName     = "testdb"
)

type MyCustomClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type User struct {
	Email   string
	First   string
	Last    string
	City    string
	Country string
	Age     int
}

var jwtSecret = os.Getenv("JWT_SECRET")

func getToken(req *http.Request) string {
	hdr := req.Header.Get("authorization")
	if hdr == "" {
		return ""
	}

	token := strings.Split(hdr, "Bearer ")[1]
	return token
}

func main() {
	startService()
}

func startService() {

	r := gin.New()

	db, err := sql.Open(dbDriver, fmt.Sprintf("%s:%s@/%s", dbUser, dbPassword, dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	r.GET("/", func(c *gin.Context) {

		tokenString := getToken(c.Request)
		if tokenString == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(*MyCustomClaims)

		query := "SELECT * FROM users WHERE EMAIL = ?"
		row := db.QueryRow(query, claims.Email)
		var user User
		err2 := row.Scan(&user.Email, &user.First, &user.Last, &user.City, &user.Country, &user.Age)
		if err2 != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.JSON(http.StatusOK, user)
	})

	r.Run(":3000")
}
