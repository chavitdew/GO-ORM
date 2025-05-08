package main

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret = []byte("Test") //For Test

type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	Password string `json:"password"`
}

func createUser(db *gorm.DB, user *User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	result := db.Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func loginUser(db *gorm.DB, user *User) (string, error) {
	var selectedUser User

	result := db.Where("email = ?", user.Email).First(&selectedUser)
	if result.Error != nil {

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid credentials")
		}
		return "", result.Error
	}

	// Compare the password hash
	err := bcrypt.CompareHashAndPassword([]byte(selectedUser.Password), []byte(user.Password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate a JWT token
	token, err := generateJWT(selectedUser.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func generateJWT(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
