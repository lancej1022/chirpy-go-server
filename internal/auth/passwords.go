package auth

import "golang.org/x/crypto/bcrypt"

// Hash the password using the bcrypt.GenerateFromPassword function. Bcrypt is a secure hash function that is intended for use with passwords.
func HashPassword(password string) (string, error) {
	bcrypt.GenerateFromPassword()
	// TODO:
	return "", nil
}
