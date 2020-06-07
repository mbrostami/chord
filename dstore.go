package chord

import (
	"encoding/json"
	"time"

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

type Record struct {
	CreationTime time.Time `json:"creation_time"`
	Content      []byte    `json:"content"`
}

func (r *Record) Hash() [helpers.HashSize]byte {
	json, _ := json.Marshal(r)
	return helpers.Hash(string(json))
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
	key := helpers.Hash(string(json))
	// log.Debugf("storing data %x: %v", key, json)
	d.db[key] = &json
	return true
}

func (d *DStore) PutRecord(record Record) bool {
	json, _ := json.Marshal(record)
	key := helpers.Hash(string(json))
	// log.Debugf("storing data %x: %v", key, json)
	d.db[key] = &json
	return true
}

func (d *DStore) GetRange(fromKey [helpers.HashSize]byte, toKey [helpers.HashSize]byte) map[[helpers.HashSize]byte]*Record {
	data := make(map[[helpers.HashSize]byte]*Record)
	for key, item := range d.db {
		if helpers.BetweenR(key, fromKey, toKey) {
			record := Record{}
			json.Unmarshal(*item, &record)
			data[key] = &record
		}
	}
	return data
}

func (d *DStore) GetAll() map[[helpers.HashSize]byte]*Record {
	data := make(map[[helpers.HashSize]byte]*Record)
	for key, item := range d.db {
		record := Record{}
		json.Unmarshal(*item, &record)
		data[key] = &record
	}
	return data
}
