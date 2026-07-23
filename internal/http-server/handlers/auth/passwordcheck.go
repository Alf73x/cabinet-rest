package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func CheckPassword(passwordHash, password string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(password),
	)

	return err == nil
}

func TestPassword() {
	hash := "..."

	password := "p5a4s3s3w4o5rd"

	err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)

	if err != nil {
		fmt.Println("Password is incorrect")
		return
	}

	fmt.Println("Password is correct")
}
