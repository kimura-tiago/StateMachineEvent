package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

func SavePayout(payout *Payout) error {
	db := getDb()
	db[payout.Id] = *payout
	writeDb(db)
	return nil
}

func GetPayout(id string) (*Payout, error) {
	db := getDb()
	if payout, ok := db[id]; ok {
		return &payout, nil
	}

	return nil, errors.New("not found")
}

func getDb() map[string]Payout {
	payoutsBytes, _ := ioutil.ReadFile("payouts.json")
	db := make(map[string]Payout)
	json.Unmarshal(payoutsBytes, &db)
	return db
}

func writeDb(db map[string]Payout) {
	bytes, _ := json.Marshal(db)

	file, _ := os.Create("payouts.json")
	defer file.Close()
	file.Write(bytes)
}
