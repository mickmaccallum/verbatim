package model

import (
	"errors"
	"regexp"
)

type pattern string

const (
	// Source: https://www.safaribooksonline.com/library/view/regular-expressions-cookbook/9780596802837/ch07s16.html
	ipv4Pattern pattern = "^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$"

	// Source: http://stackoverflow.com/a/17871737/716216
	ipv6Pattern pattern = "(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))"
)

func isValidIp(ipAddress string) (bool, error) {
	// Min length of IPv6, max length of IPv6.
	if len(ipAddress) < 3 || len(ipAddress) > 45 {
		return false, errors.New("Invalid IP Address length")
	}

	// validate IPv4 address
	match, err := regexp.MatchString(string(ipv4Pattern), ipAddress)
	if err != nil {
		return false, err
	}

	if !match {
		// validate IPv6 address
		match, err = regexp.MatchString(string(ipv6Pattern), ipAddress)
		if err != nil {
			return false, err
		}

		if !match {
			return false, errors.New("Invalid IP address")
		}
	}

	return true, nil
}
