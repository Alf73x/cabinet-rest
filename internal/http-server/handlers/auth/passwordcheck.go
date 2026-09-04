package auth

import (
	"sync"

	"golang.org/x/crypto/bcrypt"
)

var (
	dummyHashOnce sync.Once
	dummyHash     []byte
)

func CheckPassword(passwordHash, password string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(password),
	)

	return err == nil
}

func initDummyPasswordHash() {
	dummyHashOnce.Do(func() {
		dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-login-password"), bcrypt.DefaultCost)
	})
}

func CheckDummyPassword(password string) {
	initDummyPasswordHash()
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
}
