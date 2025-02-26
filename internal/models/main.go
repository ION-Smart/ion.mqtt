package models

import (
	"fmt"
	"strconv"
	"time"
)

type Modulo struct {
	CodModulo    int
	Abreviacion  string
	NombreModulo string
	CodSector    int
}

type DateTime struct {
	Time string
}

func (d *DateTime) GetDateTimeFromStringMilli(tstamp string) {
	timestamp, err := strconv.Atoi(tstamp)
	if err != nil {
		panic(err)
	}

	t := time.UnixMilli(int64(timestamp))
	d.Time = fmt.Sprintf("%4d-%02d-%02d %02d:%02d:%02d",
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
	)
}
