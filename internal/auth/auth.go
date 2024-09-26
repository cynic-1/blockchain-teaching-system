package auth

import (
	"crypto/rand"
	"fmt"
	"github.com/cynic-1/blockchain-teaching-system/internal/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

var secretKey []byte

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

func InitSecretKey() error {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return err
	}
	secretKey = bytes
	return nil
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

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer <token>"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// 返回用于验证签名的密钥
			return secretKey, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		//if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		//	// 将 claims 存储在上下文中
		//	c.Set("claims", claims)
		//	c.Next()
		//} else {
		//	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		//	c.Abort()
		//	return
		//}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}
