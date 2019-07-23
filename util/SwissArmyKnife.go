package util

import "log"

func CheckError(msg string, err error) {
	if err != nil {
		log.Fatal(msg, '\n', err)
	}
}
