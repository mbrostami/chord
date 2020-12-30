package chord

import (
	"encoding/base64"
	"encoding/json"

	"github.com/mbrostami/chord/helpers"
)

type Data struct {
	records       map[[helpers.HashSize]byte]*Record
	Base64Records map[string]*Record             `json:"records"`
	RootHash      [helpers.HashSize]byte         `json:"root"`
	Ranges        map[int][helpers.HashSize]byte `json:"ranges"`
}

func NewData(records map[[helpers.HashSize]byte]*Record, ranges map[int][helpers.HashSize]byte, root [helpers.HashSize]byte) *Data {
	d := Data{
		records:  records,
		RootHash: root,
		Ranges:   ranges,
	}
	return &d
}

func (d *Data) GetRecords() map[[helpers.HashSize]byte]*Record {
	return d.records
}

func (d *Data) GetRecord(id [helpers.HashSize]byte) *Record {
	return d.records[id]
}

func SerializeData(d *Data) ([]byte, error) {
	Records := make(map[string]*Record)
	for key, value := range d.records {
		newKey := base64.StdEncoding.EncodeToString(key[:])
		Records[newKey] = value
	}
	d.Base64Records = Records
	return json.Marshal(d)
}

func UnserializeData(jsonData []byte) *Data {
	data := Data{}
	json.Unmarshal(jsonData, &data)
	records := make(map[[helpers.HashSize]byte]*Record)
	for key, value := range data.Base64Records {
		var newKeyByte [helpers.HashSize]byte
		newKey, _ := base64.StdEncoding.DecodeString(key)
		copy(newKey, newKeyByte[:])
		records[newKeyByte] = value
	}
	data.records = records
	return &data
}
