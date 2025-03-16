package workloadgen

import (
	"strconv"

	"golang.org/x/exp/rand"
)

func GenUserHandler() (string, string) {
	id := rand.Intn(500)
	username := "Cornell_" + strconv.Itoa(id)
	password := ""
	for i := 0; i < 10; i += 1 {
		password += strconv.Itoa(id)
	}
	return username, password
}

func GenSearchHandler() (float64, float64, string, string) {
	lat := 38.0235
	lon := -122.095
	lat = lat + (float64(rand.Intn(481))-240.5)/1000.0
	lon = lon + (float64(rand.Intn(325))-157.0)/1000.0
	inDay := rand.Intn(14) + 9
	outDay := rand.Intn(24-1-inDay) + inDay + 1
	var inDate, outDate string
	if inDay > 9 {
		inDate = "2015-04-" + strconv.Itoa(inDay)
	} else {
		inDate = "2015-04-0" + strconv.Itoa(inDay)
	}
	if outDay > 9 {
		outDate = "2015-04-" + strconv.Itoa(outDay)
	} else {
		outDate = "2015-04-0" + strconv.Itoa(outDay)
	}
	return lat, lon, inDate, outDate
}

func GenRecommendHandler() (float64, float64, string) {
	lat := 38.0235
	lon := -122.095
	lat = lat + (float64(rand.Intn(481))-240.5)/1000.0
	lon = lon + (float64(rand.Intn(325))-157.0)/1000.0
	req := ""
	rnum := rand.Intn(100)
	if rnum < 33 {
		req = "dis"
	} else if rnum < 66 {
		req = "rate"
	} else {
		req = "price"
	}
	return lat, lon, req
}

func GenReservationHandler() (string, string, string, string, string, string, int64) {
	inDay := rand.Intn(14) + 9
	outDay := rand.Intn(24-1-inDay) + inDay + 1
	var inDate, outDate string
	if inDay > 9 {
		inDate = "2015-04-" + strconv.Itoa(inDay)
	} else {
		inDate = "2015-04-0" + strconv.Itoa(inDay)
	}
	if outDay > 9 {
		outDate = "2015-04-" + strconv.Itoa(outDay)
	} else {
		outDate = "2015-04-0" + strconv.Itoa(outDay)
	}
	hotelid := rand.Intn(80) + 1
	hotelId := strconv.Itoa(hotelid)
	id := rand.Intn(500)
	username := "Cornell_" + strconv.Itoa(id)
	password := ""
	for i := 0; i < 10; i += 1 {
		password += strconv.Itoa(id)
	}
	customerName := username
	roomNUmber := 1
	return inDate, outDate, hotelId, username, password, customerName, int64(roomNUmber)
}
