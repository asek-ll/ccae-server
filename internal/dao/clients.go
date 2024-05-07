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

func NewClientsDao(db *sql.DB) (*ClientsDao, error) {

	sqlStmt := `

	CREATE TABLE IF NOT EXISTS clients (
		id string NOT NULL PRIMARY KEY,
		role string NOT NULL,
		online bool NOT NULL,
		last_login timestamp,
		wsclient_id integer
	);

	UPDATE clients SET online = false;
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &ClientsDao{db: db}, nil
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

func (c *ClientsDao) GetOnlineClientIdOfType(clientType string) (uint, error) {
	row := c.db.QueryRow("SELECT wsclient_id FROM clients WHERE online = true AND role = ?", clientType)
	err := row.Err()
	if err != nil {
		return 0, err
	}

	var id int
	row.Scan(&id)

	return uint(id), nil
}

func (c *ClientsDao) LoginClient(clientId string, role string, wsclientId uint) error {
	stmt, err := c.db.Prepare("INSERT OR REPLACE INTO clients (id, role, online, last_login, wsclient_id) VALUES (?, ?, true, datetime(), ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(clientId, role, wsclientId)
	return err
}

func (c *ClientsDao) LogoutClient(clientId string) error {
	stmt, err := c.db.Prepare("UPDATE clients SET online = false, wsclient_id = NULL WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(clientId)
	return err
}
