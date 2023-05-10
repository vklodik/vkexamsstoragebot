package database

import (
	"database/sql"
	"fmt"
)

func Set(db *sql.DB, userId int, serviceName, login, password string) error {
	query := `INSERT INTO users(user_id, service_name, login, password) VALUES($1, $2, $3, $4)`

	_, err := db.Exec(query, userId, serviceName, login, password)
	if err != nil {
		return err
	}

	return nil
}

func Get(db *sql.DB, userId int, serviceName string) (string, string, error) {
	query := `SELECT login, password FROM users WHERE user_id=$1 AND service_name=$2`

	row := db.QueryRow(query, userId, serviceName)

	var login, password string
	err := row.Scan(&login, &password)
	if err != nil {
		return "", "", err
	}

	return login, password, nil
}

func Del(db *sql.DB, userId int, serviceName string) error {
	query := `DELETE FROM users WHERE user_id=$1 AND service_name=$2`

	res, err := db.Exec(query, userId, serviceName)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no rows were affected")
	}

	return nil
}

func GetServices(db *sql.DB, userId int) ([]string, error) {
	query := `SELECT service_name FROM users WHERE user_id=$1`

	rows, err := db.Query(query, userId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var serviceNames []string
	for rows.Next() {
		var serviceName string
		err := rows.Scan(&serviceName)
		if err != nil {
			return nil, err
		}
		serviceNames = append(serviceNames, serviceName)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return serviceNames, nil
}
