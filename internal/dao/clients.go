package dao

import (
	"database/sql"
	"time"
)

type Client struct {
	ID         int
	Label      string
	Role       string
	Online     bool
	LastLogin  time.Time
	WSClientID uint
}

type ClientsDao struct {
	db *sql.DB
}

func NewClientsDao(db *sql.DB) (*ClientsDao, error) {

	sqlStmt := `

	CREATE TABLE IF NOT EXISTS clients (
		id int NOT NULL PRIMARY KEY,
		label string,
		role string NULL,
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
	rows, err := c.db.Query("select id, label, role, online, last_login, wsclient_id from clients")
	if err != nil {
		return nil, err
	}

	return readClients(rows)
}

func readClients(rows *sql.Rows) ([]Client, error) {
	defer rows.Close()

	var result []Client
	for rows.Next() {
		var id int
		var label string
		var role string
		var online bool
		var lastLogin time.Time
		var wsClientID uint

		err := rows.Scan(&id, &label, &role, &online, &lastLogin, &wsClientID)
		if err != nil {
			return nil, err
		}
		result = append(result, Client{
			ID:         id,
			Label:      label,
			Role:       role,
			Online:     online,
			LastLogin:  lastLogin,
			WSClientID: wsClientID,
		})
	}
	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// func (c *ClientsDao) GetOnlineClientIdOfType(clientType string) (uint, error) {
// 	row := c.db.QueryRow("SELECT wsclient_id FROM clients WHERE online = true AND role = ?", clientType)
// 	err := row.Err()
// 	if err != nil {
// 		return 0, err
// 	}

// 	var id int
// 	row.Scan(&id)

// 	return uint(id), nil
// }

func (c *ClientsDao) GetClientByID(id int) (*Client, error) {
	rows, err := c.db.Query("SELECT id, label, role, online, last_login, wsclient_id FROM clients WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	clients, err := readClients(rows)
	if err != nil {
		return nil, err
	}

	if len(clients) == 0 {
		return nil, nil
	}

	return &clients[0], nil
}

func (c *ClientsDao) CreateClient(client *Client) error {
	_, err := c.db.Exec(`
		INSERT INTO clients (id, label, role, online, last_login, wsclient_id)
		VALUES (?, ?, ?, ?, ?, ?)
		`, client.ID, client.Label, client.Role, client.Online, client.LastLogin, client.WSClientID)
	return err
}

func (c *ClientsDao) LoginClient(client *Client, wsClientID uint) error {
	client.Online = true
	client.LastLogin = time.Now()
	client.WSClientID = wsClientID
	_, err := c.db.Exec(`
		UPDATE clients
		SET online = ?, last_login = ?, wsclient_id = ?
		WHERE id = ?
		`, client.Online, client.LastLogin, client.WSClientID, client.ID)
	return err
}

func (c *ClientsDao) LogoutClient(wsClientId uint) error {
	_, err := c.db.Exec("UPDATE clients SET online = false, wsclient_id = NULL WHERE wsclient_id = ?", wsClientId)
	return err
}
