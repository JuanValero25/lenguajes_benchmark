package main

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"net/http"
	"runtime"
	"runtime/debug"
	"sync"
	"testing"
	"time"
)

func BenchmarkGetValue(b *testing.B) {
	go startService()
	time.Sleep(100 * time.Millisecond)
	httpCli := http.Client{}
	strings, err := generateJWTToken("aaATrsAeltmYkYuIvgZSlEZuBYwFCCsbTbUIDiFwEsOTAuwSFx@example.com")

	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodGet, "http://localhost:3000", nil)
	if err != nil {
		return
	}

	req.Header.Set("authorization", strings)

	b.Run("some benchmark", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			httpCli.Do(req)
		}
	})
}

func BenchmarkGolang(b *testing.B) {
	go startService()
	time.Sleep(100 * time.Millisecond)
	debug.SetGCPercent(-1)

	jwtTokens := GetAllEmailsToJWT()
	runtime.GC()
	b.Run("10_connections", func(b *testing.B) {
		benchmarkWithNumberOfConnection("10_connections_golang", "3000", 10, 50000, jwtTokens)
	})

	runtime.GC()
	time.Sleep(2 * time.Millisecond)
	b.Run("50_connections", func(b *testing.B) {
		benchmarkWithNumberOfConnection("50_connections_golang", "3000", 50, 50000, jwtTokens)
	})
	runtime.GC()
	time.Sleep(2 * time.Millisecond)
	b.Run("100_connections", func(b *testing.B) {
		benchmarkWithNumberOfConnection("100_connections_golang", "3000", 100, 50000, jwtTokens)
	})

}

func BenchmarkRust(b *testing.B) {
	debug.SetGCPercent(-1)

	jwtTokens := GetAllEmailsToJWT()
	runtime.GC()
	b.Run("10_connections", func(b *testing.B) {
		benchmarkWithNumberOfConnection("10_connections_rust", "3001", 10, 50000, jwtTokens)
	})

	runtime.GC()
	time.Sleep(2 * time.Millisecond)
	b.Run("50_connections", func(b *testing.B) {
		benchmarkWithNumberOfConnection("50_connections_rust", "3001", 50, 50000, jwtTokens)
	})
	runtime.GC()
	time.Sleep(2 * time.Millisecond)
	b.Run("100_connections", func(b *testing.B) {
		benchmarkWithNumberOfConnection("100_connections_rust", "3001", 100, 50000, jwtTokens)
	})

}

func benchmarkWithNumberOfConnection(description, portService string, numberOfConnection, nroOfRequest int, jwtTokens []string) (fails []error) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, numberOfConnection)

	tokenIndex := 0
	httpCli := &http.Client{Transport: &http.Transport{
		ResponseHeaderTimeout: time.Hour,
		MaxConnsPerHost:       1500,
		MaxIdleConns:          0,
		MaxIdleConnsPerHost:   0,
		ForceAttemptHTTP2:     true,
	}}

	starTime := time.Now()
	wg.Add(nroOfRequest)
	for i := 0; i < nroOfRequest; i++ {
		tokenIndex++
		if tokenIndex == len(jwtTokens)-1 {
			tokenIndex = 0
		}

		semaphore <- struct{}{}
		go func() {
			makeRequest(httpCli, portService, jwtTokens[tokenIndex])
			<-semaphore // Liberar el semáforo cuando la goroutine ha terminado
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("\n test %v request finished in %v ", description, time.Since(starTime))
	return fails
}

func makeRequest(client *http.Client, portService, jwtToken string) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:"+portService, nil)
	if err != nil {
		fmt.Printf("error creando el request")
		return
	}
	req.Header.Set("authorization", jwtToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error haciendo el request el request")
		return
	}
	defer resp.Body.Close()

	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		fmt.Println("Error al leer la respuesta:", err)
		return
	}

}

func generateJWTToken(email string) (string, error) {
	// Crear los claims del token
	claims := jwt.MapClaims{
		"email": email,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Expira en 24 horas
	}

	// Crear el token JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmar el token con la clave secreta
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return "Bearer " + tokenString, nil
}
