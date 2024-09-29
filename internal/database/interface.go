// database/interface.go
package database

import "github.com/cynic-1/blockchain-teaching-system/internal/models"

type DatabaseInterface interface {
	Close() error
	SaveUser(user *models.User) error
	GetUser(userID string) (*models.User, error)
}
