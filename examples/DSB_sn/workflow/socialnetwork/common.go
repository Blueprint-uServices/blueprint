package socialnetwork

import (
	"bytes"
	"net"
	"strconv"
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

			return strconv.FormatUint(mac, 16)
		}
	}

	return "0"
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
