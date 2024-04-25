package dao

import (
	"database/sql"
)

type Client struct {
	Id string
}

type ClientsDao struct {
	db *sql.DB
}

func (c *ClientsDao) GetClients() ([]Client, error) {
	rows, err := c.db.Query("select id from clients")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Client
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		result = append(result, Client{Id: id})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}
