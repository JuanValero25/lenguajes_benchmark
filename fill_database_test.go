package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

// Función para generar un usuario aleatorio
func generateUser() User {
	email := generateRandomString(50) + "@example.com"
	first := generateRandomString(10)
	last := generateRandomString(10)
	city := generateRandomString(8)
	country := generateRandomString(8)
	age := rand.Intn(50) + 18 // Generar una edad entre 18 y 67 años

	return User{
		Email:   email,
		First:   first,
		Last:    last,
		City:    city,
		Country: country,
		Age:     age,
	}
}

// Función para generar una cadena aleatoria de longitud n
func generateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

// Función para insertar un usuario en la base de datos
func insertUser(db *sql.DB, user User) {
	query := "INSERT INTO users (email, first, last, city, country, age) VALUES (?, ?, ?, ?, ?, ?)"

	_, err := db.Exec(query, user.Email, user.First, user.Last, user.City, user.Country, user.Age)
	if err != nil {
		user.Email = generateRandomString(5) + user.Email
		insertUser(db, user)
		log.Println(err)
	}
}

func TestGetEmail(t *testing.T) {
	GetAllEmailsToJWT()
}

func GetAllEmailsToJWT() []string {
	//"root:dbpwd@tcp(localhost:3306)/testdb"
	//	dbDriver   = "mysql"
	//	dbUser     = "root"
	//	dbPassword = "dbpwd"
	//	dbName     = "testdb"
	//)
	db, err := sql.Open(dbDriver, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("SELECT email FROM users")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// Crea un mapa para almacenar los valores de email
	emails := make([]string, 0, 100000)

	// Recorre los resultados y llena el mapa
	for rows.Next() {
		var email string
		err := rows.Scan(&email)
		if err != nil {
			panic(err)
		}

		jetToken, err := generateJWTToken(email)

		if err != nil {
			return nil
		}

		emails = append(emails, jetToken)
	}

	// Maneja cualquier error de iteración
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return emails
}

func TestFillDataBase(t *testing.T) {
	t.Log(fmt.Sprintf("%s:%s@/%s", dbUser, dbPassword, dbName))
	db, err := sql.Open(dbDriver, fmt.Sprintf("%s:%s@/%s", dbUser, dbPassword, dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // max 5 query at the same time
	// Generar 100,000 usuarios y guardarlos en la base de datos
	wg.Add(100000)
	for i := 0; i < 100000; i++ {
		user := generateUser()
		semaphore <- struct{}{}
		go func() {
			insertUser(db, user)
			wg.Done()
			<-semaphore
		}()

	}
	wg.Wait()
	elapsed := time.Since(start)
	log.Printf("Tiempo transcurrido: %s", elapsed)
}
