package BLC

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
)

func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	/*
		big endian：最高字节在地址最低位，最低字节在地址最高位，依次排列。
		little endian：最低字节在最低位，最高字节在最高位，反序排列。
	*/
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes() // [0 0 0 0 0 0 1 0]
}

func JSONToArray(jsonString string) []string {

	//json 到 []string
	var sArr []string
	if err := json.Unmarshal([]byte(jsonString), &sArr); err != nil {
		log.Panic(err)
	}
	return sArr
}

func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func gobEncodeWithRegister(data interface{}, inter interface{}) []byte {
	var buff bytes.Buffer
	gob.Register(inter)
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func gobDecode(blockBytes []byte, o interface{}) {
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(o)
	if err != nil {
		log.Panic(err)
	}
}

func commandToBytes(command string) []byte {
	var bytes [COMMAND_LENGTH]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}

func bytesToCommand(commandBtyes []byte) string {
	var command []byte

	for _, b := range commandBtyes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}
