package BLC

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
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


// Concert json(string format) to []string
func JSONToArray(jsonString string) []string {

	//json 到 []string
	var sArr []string
	if err := json.Unmarshal([]byte(jsonString), &sArr); err != nil {
		log.Panic(err)
	}
	return sArr
}