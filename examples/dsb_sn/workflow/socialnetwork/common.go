package socialnetwork

import (
	"bytes"
	"encoding/json"
	"net"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// From: https://gist.github.com/tsilvers/085c5f39430ced605d970094edf167ba
func GetMachineID() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "0"
	}

	for _, i := range interfaces {
		if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {

			// Skip locally administered addresses
			if i.HardwareAddr[0]&2 == 2 {
				continue
			}

			var mac uint64
			for j, b := range i.HardwareAddr {
				if j >= 8 {
					break
				}
				mac <<= 8
				mac += uint64(b)
			}

			// Convert from uint64 to uint16
			mac_ui16 := uint16(mac)
			return strconv.FormatUint(uint64(mac_ui16), 16)
		}
	}

	return "0"
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Converts a json-encoded string to a bson.D document to be used as arguments for one of the other NoSQLDatabase functions
func parseNoSQLDBQuery(query string) (bson.D, error) {
	return handleFormats(query)
}

func lower(f interface{}) interface{} {
	switch f := f.(type) {
	case []interface{}:
		for i := range f {
			f[i] = lower(f[i])
		}
		return f
	case map[string]interface{}:
		lf := make(map[string]interface{}, len(f))
		for k, v := range f {
			if k == "$elemMatch" {
				lf[k] = lower(v)
			} else {
				lf[strings.ToLower(k)] = lower(v)
			}
		}
		return lf
	default:
		return f
	}
}

func handleFormats(jsonQuery string) (bdoc bson.D, err error) {

	if jsonQuery == "" {
		bdoc = bson.D{}
		return
	}

	var f interface{}
	err = json.Unmarshal([]byte(jsonQuery), &f)
	if err != nil {
		return
	}

	f = lower(f)

	lowerQuery, err := json.Marshal(f)
	if err != nil {
		return
	}
	err = bson.UnmarshalExtJSON(lowerQuery, true, &bdoc)
	return
}
