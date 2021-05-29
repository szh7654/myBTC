package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"log"
)

// IntToHex Convert int64 to byte[]
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	if err := binary.Write(buff, binary.BigEndian, num); err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func ToBytes(data interface{}) []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return res.Bytes()
}