package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	switch os.Args[1] {
	case "hash":
		hash(os.Args[2])
	case "compare":
		compare(os.Args[2], os.Args[3])
	default:
		fmt.Printf("Invalid command: %v\n", os.Args[1])
	}
}

/* go run cmd/bcrypt/bcrypt.go hash "secret password" */
func hash(password string) {
	hashedBytes, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("error hashing: %v\n", err)
		return
	}
	hash := string(hashedBytes)
	fmt.Println(hash)
}

/*
	go run cmd/bcrypt/bcrypt.go compare \

"secret password" \
'$2a$10$GIdKwhQ8jUnH8bNS/CDDM.yhb9tnPc6sZWI4NIWytGOvGwGAHIJLO'
*/
func compare(password, hash string) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Printf("Password is invalid: %v\n", password)
		return
	}
	fmt.Println("Password is correct!")
}
