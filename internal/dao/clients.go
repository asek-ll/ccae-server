package dao

import (
	"database/sql"
	"time"
)

type Client struct {
	ID         string
	Label      string
	Role       string
	Online     bool
	LastLogin  time.Time
	WSClientID *uint
	Authorized bool
}

type ClientsDao struct {
	db *sql.DB
}

func NewClientsDao(db *sql.DB) (*ClientsDao, error) {

	sqlStmt := `

	CREATE TABLE IF NOT EXISTS clients (
		id string NOT NULL PRIMARY KEY,
		label string,
		role string,
		online bool NOT NULL,
		last_login timestamp,
		wsclient_id integer,
		authorized bool NOT NULL
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
	rows, err := c.db.Query("select id, label, role, online, last_login, wsclient_id, authorized from clients")
	if err != nil {
		return nil, err
	}

	return readClients(rows)
}

func readClients(rows *sql.Rows) ([]Client, error) {
	defer rows.Close()

	var result []Client
	for rows.Next() {
		var id string
		var label string
		var role string
		var online bool
		var lastLogin time.Time
		var wsClientID *uint
		var authorized bool

		err := rows.Scan(&id, &label, &role, &online, &lastLogin, &wsClientID, &authorized)
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
			Authorized: authorized,
		})
	}
	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *ClientsDao) GetClientByID(id string) (*Client, error) {
	rows, err := c.db.Query("SELECT id, label, role, online, last_login, wsclient_id, authorized FROM clients WHERE id = ?", id)
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
		INSERT INTO clients (id, label, role, online, last_login, wsclient_id, authorized)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		`, client.ID, client.Label, client.Role, client.Online, client.LastLogin, client.WSClientID, client.Authorized)
	return err
}

func (c *ClientsDao) UpdateClient(client *Client) error {
	_, err := c.db.Exec(`
		UPDATE clients
		SET label = ?, role = ?
		WHERE id = ?
		`, client.Label, client.Role, client.ID)
	return err
}

func (c *ClientsDao) LoginClient(client *Client, wsClientID uint) error {
	client.Online = true
	client.LastLogin = time.Now()
	client.WSClientID = &wsClientID
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

func (c *ClientsDao) AuthorizeClient(client *Client) error {
	client.Authorized = true
	_, err := c.db.Exec(`
		UPDATE clients
		SET authorized = ?
		WHERE id = ?
		`, client.Authorized, client.ID)
	return err
}

func (c *ClientsDao) DeleteClient(client *Client) error {
	_, err := c.db.Exec("DELETE FROM clients WHERE id = ?", client.ID)
	return err
}
