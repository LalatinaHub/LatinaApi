package member

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"

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

func GetMember(id any) (int, string) {
	var (
		expired         int    = 1
		password, query string = "", ""
	)

	if reflect.TypeOf(id).Kind() == reflect.String {
		query = fmt.Sprintf(`SELECT EXTRACT(DAY FROM NOW() - (SELECT EXPIRED FROM users WHERE PASSWORD = '%s')), PASSWORD FROM users WHERE PASSWORD = '%s'`, id, id)
	} else {
		query = fmt.Sprintf(`SELECT EXTRACT(DAY FROM NOW() - (SELECT EXPIRED FROM users WHERE ID = %d)), PASSWORD FROM users WHERE ID = %d`, id, id)
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
	var query string

	if IsExists(id) {
		query = fmt.Sprintf(`UPDATE users SET EXPIRED = NOW() + INTERVAL '%d MONTH' WHERE ID = %d`, subs, id)
	} else {
		hash := GenerateHash(strconv.FormatInt(id, 10))
		query = fmt.Sprintf(`INSERT INTO users (ID, EXPIRED, PASSWORD) VALUES (%d, NOW() + INTERVAL '%d MONTH', '%s')`, id, subs, hash)
	}

	_, err := db.New().Conn().Exec(query)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func ChangePassword(id int64, password string) bool {
	if _, isExists := GetMember(password); isExists != "" {
		return false
	}

	query := fmt.Sprintf(`UPDATE users SET PASSWORD = '%s' WHERE ID = %d`, password, id)
	_, err := db.New().Conn().Exec(query)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func GenerateHash(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}
