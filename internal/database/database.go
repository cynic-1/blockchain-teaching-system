package database

import (
	"encoding/json"
	"github.com/cynic-1/blockchain-teaching-system/internal/models"
	"github.com/dgraph-io/badger/v3"
)

type Database struct {
	db *badger.DB
}

func NewDatabase(path string) (*Database, error) {
	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) SaveUser(user *models.User) error {
	return d.db.Update(func(txn *badger.Txn) error {
		userBytes, err := json.Marshal(user)
		if err != nil {
			return err
		}
		return txn.Set([]byte(user.ID), userBytes)
	})
}

func (d *Database) GetUser(userID string) (*models.User, error) {
	var user models.User
	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(userID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &user)
		})
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}
