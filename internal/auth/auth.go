package auth

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/cynic-1/blockchain-teaching-system/internal/models"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var secretKey string

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidateUser(user *models.User, password string) bool {
	return CheckPasswordHash(password, user.Password)
}

func generateSecretKey(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func InitSecretKey() error {
	var err error
	secretKey, err = generateSecretKey(32)
	return err
}

// 创建token
func CreateToken(userID string) (string, error) {
	var err error
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	//atClaims["authorized"] = true
	atClaims["user_id"] = userID

	// expire after 100 mins
	atClaims["exp"] = time.Now().Add(time.Minute * 100).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

// 验证token
func VerifyTokenAndExtractUserID(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secretKey, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", jwt.ErrSignatureInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrSignatureInvalid
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", jwt.ErrSignatureInvalid
	}

	return userID, nil
}
