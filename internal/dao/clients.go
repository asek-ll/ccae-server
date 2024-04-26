package dao

import (
	"database/sql"
	"time"
)

type Client struct {
	Id        string
	Role      string
	Online    bool
	LastLogin time.Time
}

type ClientsDao struct {
	db *sql.DB
}

func (c *ClientsDao) GetClients() ([]Client, error) {
	rows, err := c.db.Query("select id, role, online, last_login from clients")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Client
	for rows.Next() {
		var id string
		var role string
		var online bool
		var lastLogin time.Time

		err = rows.Scan(&id, &role, &online, &lastLogin)
		if err != nil {
			return nil, err
		}
		result = append(result, Client{
			Id:        id,
			Role:      role,
			Online:    online,
			LastLogin: lastLogin,
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *ClientsDao) LoginClient(clientId string, role string) error {
	stmt, err := c.db.Prepare("INSERT OR REPLACE INTO clients (id, role, online, last_login) VALUES (?, ?, true, datetime())")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(clientId, role)
	return err
}

func (c *ClientsDao) LogoutClient(clientId string) error {
	stmt, err := c.db.Prepare("UPDATE clients SET online = false WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(clientId)
	return err
}
