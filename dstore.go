package chord

import (
	log "github.com/sirupsen/logrus"

	"github.com/mbrostami/chord/helpers"
)

// DBREPLICAS this must be less than predecessor list count
const DBREPLICAS int = 1

type DStore struct {
	db map[[helpers.HashSize]byte]*[]byte
}

func NewDStore() *DStore {
	return &DStore{
		db: make(map[[helpers.HashSize]byte]*[]byte),
	}
}

func (d *DStore) Put(key [helpers.HashSize]byte, value []byte) bool {
	log.Debugf("storing data %x: %v", key, value)
	d.db[key] = &value
	return true
}

func (d *DStore) GetRange(fromKey [helpers.HashSize]byte, toKey [helpers.HashSize]byte) map[[helpers.HashSize]byte]*[]byte {
	data := make(map[[helpers.HashSize]byte]*[]byte)
	for key, item := range d.db {
		if helpers.BetweenR(key, fromKey, toKey) {
			data[key] = item
		}
	}
	return data
}
