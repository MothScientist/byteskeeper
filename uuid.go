package main

import (
	"time"

	"github.com/google/uuid"
)

var SecondsBetween15oct1582andStartUNIX int64 = 12_219_292_800

func GetDateFromUUIDv7(uuid uuid.UUID) time.Time {
	secondsFrom15oct1582 := int64(uuid.Time() / 10_000_000) // 100*ns -> 1*sec
	secondsFromUnix := secondsFrom15oct1582 - SecondsBetween15oct1582andStartUNIX
	return time.Unix(secondsFromUnix, 0)
}
