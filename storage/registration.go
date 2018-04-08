package storage

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

//go:generate fungen -package storage -types Registration

type Registration struct {
	Subscriber   string `json:"subscriber"`
	Subscription string `json:"subscription"`
	Created      int64  `json:"created,omitempty"`
	Updated      int64  `json:"created,omitempty"`
}

type DB struct {
	filename string
	mutex    *sync.Mutex
}

var gRegistrations RegistrationList
var gMutex = &sync.Mutex{}

// Open the JSON database storage
func Open(filename string) *DB {
	db := &DB{filename: filename, mutex: &sync.Mutex{}}
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		gRegistrations = make(RegistrationList, 0)
		return db
	}

	jsonString, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("notification: Unable to read the registrations file %s [%+v]", filename, err)
	}
	err = json.Unmarshal(jsonString, &gRegistrations)
	if err != nil {
		log.Fatalf("notification: Unable to deserialize the registrations file %s [%+v]", filename, err)
	}
	return db
}

func (db *DB) FindRegistration(subscriber string) RegistrationList {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	registrations := gRegistrations.Filter(func(registration Registration) bool {
		return subscriber == registration.Subscriber
	})
	return registrations
}

func (db *DB) Filter(filter func(registration Registration) bool) RegistrationList {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return gRegistrations.Filter(filter)
}

func (db *DB) Register(registration Registration) *DB {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	now := time.Now().UnixNano()
	if registration.Created == 0 {
		registration.Created = now
	}
	registration.Updated = now
	gRegistrations = append(gRegistrations, registration)
	return db
}

func (db *DB) Unregister(registration Registration) *DB {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	gRegistrations = gRegistrations.Filter(func(aRegistration Registration) bool {
		return aRegistration.Subscriber != registration.Subscriber ||
			aRegistration.Subscription != registration.Subscription
	})
	return db
}

func (db *DB) WriteRegistrations() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	jsonString, err := json.Marshal(&gRegistrations)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(db.filename, jsonString, 0644)
}

func (RegistrationList) Length() int {
	return len(gRegistrations)
}
