package chord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mbrostami/chord/helpers"
	bolt "go.etcd.io/bbolt"
)

// DBREPLICAS this must be less than predecessor list count
const DBREPLICAS int = 1
const bucket string = "storage"

type DStore struct {
	database *bolt.DB
	db       map[[helpers.HashSize]byte]*[]byte
}

func NewDStore() *DStore {
	db, err := bolt.Open("/tmp/chord_"+string(time.Now().Second()), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucket))
		if err != nil {
			fmt.Printf("create bucket: %s", err)
		}
		return nil
	})
	return &DStore{
		database: db,
		db:       make(map[[helpers.HashSize]byte]*[]byte),
	}
}

type Record struct {
	CreationTime time.Time              `json:"creation_time"`
	Content      []byte                 `json:"content"`
	Identifier   [helpers.HashSize]byte `json:"identifier"`
}

func (r *Record) Hash() [helpers.HashSize]byte {
	return r.Identifier
}

func (r *Record) GetJson() []byte {
	json, _ := json.Marshal(r)
	return json
}

func (d *DStore) Put(value []byte) bool {
	record := Record{
		CreationTime: time.Now(),
		Content:      value,
	}
	json, _ := json.Marshal(record)
	key := helpers.Hash(string(value))
	// log.Debugf("storing data %x: %v", key, json)
	d.database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put(key[:], json)
		return err
	})
	return true
}

func (d *DStore) PutRecord(record Record) bool {
	json, _ := json.Marshal(record)
	key := record.Identifier
	// log.Debugf("storing data %x: %v", key, json)
	d.database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put(key[:], json)
		return err
	})
	return true
}

func (d *DStore) GetRange(fromKey [helpers.HashSize]byte, toKey [helpers.HashSize]byte) map[[helpers.HashSize]byte]*Record {
	data := make(map[[helpers.HashSize]byte]*Record)
	d.database.View(func(tx *bolt.Tx) error {
		// Assume our events bucket exists and has RFC3339 encoded time keys.
		c := tx.Bucket([]byte(bucket)).Cursor()
		// Our time range spans the 90's decade.
		min := fromKey[:]
		max := toKey[:]
		// Iterate over the 90's.
		for k, value := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, value = c.Next() {
			var key [helpers.HashSize]byte
			copy(key[:helpers.HashSize], k[:helpers.HashSize])
			record := Record{}
			json.Unmarshal(value, &record)
			data[key] = &record
		}
		return nil
	})
	return data
}

func (d *DStore) Get(key [helpers.HashSize]byte) []byte {
	var result []byte
	d.database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		result = b.Get(key[:])
		return nil
	})
	return result
}

func (d *DStore) GetAll() map[[helpers.HashSize]byte]*Record {
	data := make(map[[helpers.HashSize]byte]*Record)
	d.database.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(bucket))
		c := b.Cursor()
		for k, value := c.First(); k != nil; k, value = c.Next() {
			var key [helpers.HashSize]byte
			copy(key[:helpers.HashSize], k[:helpers.HashSize])
			record := Record{}
			json.Unmarshal(value, &record)
			data[key] = &record
		}
		return nil
	})
	return data
}
