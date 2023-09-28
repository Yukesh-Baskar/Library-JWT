package helpers

import "golang.org/x/crypto/bcrypt"

type Errors struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e Errors) HandleError() *Errors {
	return &Errors{
		Message: e.Message,
		Status:  e.Status,
	}
}

func HashPassword(password string) (string, error) {
	encryptedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(encryptedPass), nil
}

func ValidatePassword(hash, pass string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass)); err != nil {
		return false
	}
	return true
}
