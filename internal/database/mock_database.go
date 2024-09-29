package database

import (
	"github.com/cynic-1/blockchain-teaching-system/internal/models"
	"github.com/dgraph-io/badger/v3"
)

type MockDatabase struct {
	Users map[string]*models.User
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		Users: make(map[string]*models.User),
	}
}

func (m *MockDatabase) Close() error {
	// 模拟关闭操作
	return nil
}

func (m *MockDatabase) SaveUser(user *models.User) error {
	m.Users[user.ID] = user
	return nil
}

func (m *MockDatabase) GetUser(userID string) (*models.User, error) {
	user, exists := m.Users[userID]
	if !exists {
		return nil, badger.ErrKeyNotFound
	}
	return user, nil
}
