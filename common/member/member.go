package member

import (
	"fmt"
	"reflect"

	"github.com/LalatinaHub/LatinaSub-go/db"
)

func IsExists(id int64) bool {
	var isExists bool = false

	query := fmt.Sprintf(`SELECT EXISTS(select * FROM users WHERE ID = %d)`, id)
	rows, err := db.New().Conn().Query(query)
	if err != nil {
		return isExists
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&isExists)
	}

	return isExists
}

func GetMember(cred any) (int, string) {
	var (
		expired         int    = 1
		password, query string = "", ""
	)

	if reflect.TypeOf(cred).Kind() == reflect.String {
		query = fmt.Sprintf(`SELECT EXTRACT(DAY FROM NOW() - (SELECT expired FROM users WHERE password = '%s')), password FROM users WHERE password = '%s'`, cred, cred)
	} else {
		query = fmt.Sprintf(`SELECT EXTRACT(DAY FROM NOW() - (SELECT expired FROM users WHERE id = %d)), password FROM users WHERE id = %d`, cred, cred)
	}

	rows, err := db.New().Conn().Query(query)
	if err != nil {
		return expired, password
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&expired, &password)
	}

	return expired, password
}
