package chord

import (
	"encoding/json"

	"github.com/mbrostami/chord/helpers"
)

type Data struct {
	Records  map[[helpers.HashSize]byte]*Record
	RootHash [helpers.HashSize]byte         `json:"root"`
	Ranges   map[int][helpers.HashSize]byte `json:"ranges"`
}

func NewData(records map[[helpers.HashSize]byte]*Record, ranges map[int][helpers.HashSize]byte, root [helpers.HashSize]byte) *Data {
	d := Data{
		Records:  records,
		RootHash: root,
		Ranges:   ranges,
	}
	return &d
}

func SerializeData(d *Data) ([]byte, error) {
	return json.Marshal(d)
}

func UnserializeData(jsonData []byte) *Data {
	data := Data{}
	json.Unmarshal(jsonData, &data)
	return &data
}
