package main

import (
	"log"

	"github.com/go-faker/faker/v4"
)

type FakeSage struct {
	Email string `faker:"email"`
	IP    string `faker:"ipv6"`
	Name  string `faker:"username"`
	RFC   string `faker:"date"`
	Unix  int64  `faker:"unix_time"`
	UUID  string `faker:"uuid_hyphenated"`
	Rand  uint32
}

func NewFakeSage() *FakeSage {
	fk := FakeSage{}
	err := faker.FakeData(&fk)
	if err != nil {
		log.Printf("faker.FakeData: %s", err)
	}
	return &fk
}
