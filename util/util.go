package util

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
)

func NewId() string {
	var buf [16]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

func ToMap(in interface{}) map[string]interface{} {
	switch in.(type) {
	case nil:
		return nil
	case map[string]interface{}:
		return in.(map[string]interface{})
	case []interface{}:
		log.Panicf("Slice is not supported. Use ToSlice function")
	}
	j, err := json.Marshal(in)
	if err != nil {
		log.Panicf("Json marshal error: %v", err.Error())
	}

	var m map[string]interface{}
	err = json.Unmarshal(j, &m)
	if err != nil {
		log.Panicf("Json unmarshal error: %v", err.Error())
	}
	return m
}

func ToStruct(in interface{}, out interface{}) {
	if in == nil {
		return
	}
	tmp, err := json.Marshal(in)
	if err != nil {
		log.Panicf("Json marshal error: %v", err.Error())
	}
	err = json.Unmarshal(tmp, &out)
	if err != nil {
		log.Panicf("Json unmarshal error: %v; json: %s", err.Error(), tmp)
	}
	return
}

func ToJson(in interface{}) (res string) {
	if in == nil {
		return ""
	}
	bts, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		log.Panicf("Json marshal error: %v", err.Error())
	}
	return string(bts)
}
