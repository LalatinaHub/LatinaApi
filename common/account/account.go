package account

import (
	"database/sql"
	"regexp"

	D "github.com/LalatinaHub/LatinaSub-go/db"
)

func Get(filter string) []D.DBScheme {
	var (
		db       = D.New()
		accounts = db.Get(filter)
		result   = []D.DBScheme{}
	)

	rows, err := db.Conn().Query("SELECT regex FROM excludes;")
	if err != nil {
		return []D.DBScheme{}
	}

	for rows.Next() {
		var domainRegex sql.NullString
		rows.Scan(&domainRegex)

		for i := 0; i < len(accounts); i++ {
			var (
				sh = []string{accounts[i].Server, accounts[i].Host}
			)

			for _, text := range sh {
				if match, _ := regexp.MatchString(domainRegex.String, text); match {
					accounts[i] = D.DBScheme{}
					break
				}
			}
		}
	}

	for _, account := range accounts {
		if account.VPN != "" {
			result = append(result, account)
		}
	}

	return result
}
