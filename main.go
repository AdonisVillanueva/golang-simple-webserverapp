package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type RegistrationDetails struct {
	Name         string
	Account      string
	Email        string
	Gender       string
	PasswordHash string
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		fmt.Fprintf(w, "Error parsing form err: %v", err)
		return
	}
	fmt.Fprintf(w, "Form: %+v\n", r.Form)

	hash, err := HashPassword(r.FormValue("password"))
	if err != nil {
		fmt.Fprintf(w, "Error when hashing password - Error: %s\n", err)
		return
	}

	registration := RegistrationDetails{
		Account:      r.FormValue("account"),
		Gender:       r.FormValue("gender"),
		Name:         r.FormValue("name"),
		Email:        r.FormValue("email"),
		PasswordHash: hash,
	}
	fmt.Fprintf(w, "Registration details: %+v\n", registration)

	match := CheckPasswordHash(r.FormValue("password"), hash)
	fmt.Fprintf(w, "Password Match hash: %t\n", match)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hello" {
		http.Error(w, "Path not found", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return

	}

	fmt.Fprintf(w, "Hello, %s", r.URL.Path)
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	m := http.NewServeMux()
	s := http.Server{Addr: ":8000", Handler: m}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
		// Cancel the context on request
		cancel()
	})
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	select {
	case <-ctx.Done():
		// Shutdown the server when the context is canceled
		s.Shutdown(ctx)
	}
	log.Printf("Finished")
}

func main() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/form", formHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/shutdown", shutdownHandler)

	fmt.Printf("Starting server at port 8080\n")
	if err := (http.ListenAndServe(":8080", nil)); err != nil {
		log.Fatal(err)
	}

}
