package storage

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type VAPID struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

func GetOrCreateVAPID(filename string) *VAPID {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		vapid := newVAPID()
		vapid.save(filename)
		return vapid
	}
	vapid := readVAPID(filename)
	return vapid
}

func newVAPID() *VAPID {
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		log.Fatalf("notification: Unable to create VAPID private/public keys [%+v]", err)
	}
	return &VAPID{Public: publicKey, Private: privateKey}
}

func readVAPID(filename string) *VAPID {
	jsonString, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("notification: Unable to read VAPID private/public keys from %s [%+v]", filename, err)
	}
	var keys VAPID
	err = json.Unmarshal(jsonString, &keys)
	if err != nil {
		log.Fatalf("notification: Unable to deserialize VAPID private/public keys from %s [%+v]", filename, err)
	}
	return &keys
}

func (keys *VAPID) save(filename string) {
	jsonString, err := json.Marshal(keys)
	if err != nil {
		log.Fatalf("notification: Unable to serialize VAPID private/public keys to %s [%+v]", filename, err)
	}
	err = ioutil.WriteFile(filename, jsonString, 0644)
	if err != nil {
		log.Fatalf("notification: Unable to write VAPID private/public keys to %s [%+v]", filename, err)
	}
}
