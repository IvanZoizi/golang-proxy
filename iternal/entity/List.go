package entity

import (
	"encoding/json"
	"os"
)

type List struct {
	Filename string   `json:"file_name"`
	Ips      []string `json:"ips"`
}

func CreateList(filename string) List {
	file, err := os.Open(filename)
	if err != nil {
		panic("File Json List not found")
	}
	defer file.Close()

	var list List
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&list)
	if err != nil {
		panic("File Json White List incorrect")
	}

	return list
}
