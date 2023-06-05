package member

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/LalatinaHub/LatinaSub-go/db"
	"github.com/sethvargo/go-password/password"
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
		query = fmt.Sprintf(`SELECT EXTRACT(DAY FROM NOW() - (SELECT EXPIRED FROM users WHERE PASSWORD = '%s')), PASSWORD FROM users WHERE PASSWORD = '%s'`, cred, cred)
	} else {
		query = fmt.Sprintf(`SELECT EXTRACT(DAY FROM NOW() - (SELECT EXPIRED FROM users WHERE ID = %d)), PASSWORD FROM users WHERE ID = %d`, cred, cred)
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

func UpdateMember(id int64, subs int) bool {
	var queries []string

	if IsExists(id) {
		queries = append(queries, fmt.Sprintf(`UPDATE users SET EXPIRED = NOW() + INTERVAL '%d MONTH' WHERE ID = %d`, subs, id))
		queries = append(queries, fmt.Sprintf("UPDATE premium SET quota = %d WHERE id = %d", 100000*subs, id))
	} else {
		password := GeneratePassword()
		queries = append(queries, fmt.Sprintf(`INSERT INTO users (ID, EXPIRED, PASSWORD) VALUES (%d, NOW() + INTERVAL '%d MONTH', '%s')`, id, subs, password))
		queries = append(queries, fmt.Sprintf(`INSERT INTO premium (ID, PASSWORD, TYPE, DOMAIN, QUOTA) VALUES (%d, default, 'dummy', 'dummy.com', 0)`, id))
	}

	_, err := db.New().Conn().Exec(fmt.Sprintf("BEGIN; %s; COMMIT;", strings.Join(queries[:], ";")))
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func ChangePassword(id int64) bool {
	password := GeneratePassword()

	query := fmt.Sprintf(`UPDATE users SET PASSWORD = '%s' WHERE ID = %d`, password, id)
	_, err := db.New().Conn().Exec(query)
	if err != nil {
		return false
	}

	return true
}

func GeneratePassword() string {
	password := password.MustGenerate(8, 2, 0, true, true)

	if _, isExists := GetMember(password); isExists != "" {
		return GeneratePassword()
	}

	return password
}
