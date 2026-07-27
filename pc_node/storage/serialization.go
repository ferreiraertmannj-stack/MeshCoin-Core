package storage

import "encoding/json"

func MarshalBlock(block interface{}) ([]byte, error) {
	return json.Marshal(block)
}

func UnmarshalBlock(data []byte, block interface{}) error {
	return json.Unmarshal(data, block)
}
