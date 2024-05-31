package dao

import "database/sql"

type ConfigOption struct {
	Key   string
	Value string
}

type ConfigsDao struct {
	db *sql.DB
}

func NewConfigsDao(db *sql.DB) (*ConfigsDao, error) {

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS configs (
		key string NOT NULL PRIMARY KEY,
		value string NOT NULL
	);

	INSERT OR IGNORE INTO configs(key, value) VALUES('do-cleanup-inventory', '');
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &ConfigsDao{db: db}, nil
}

func (d *ConfigsDao) GetConfig(key string) (string, error) {
	row := d.db.QueryRow("SELECT value FROM configs WHERE key = ?", key)
	err := row.Err()
	if err != nil {
		return "", err
	}

	var value string
	row.Scan(&value)

	return value, nil
}

func (d *ConfigsDao) SetConfig(key, value string) error {
	_, err := d.db.Exec("INSERT OR REPLACE INTO configs(key, value) VALUES(?, ?)", key, value)
	return err
}

func (d *ConfigsDao) DeleteConfig(key string) error {
	_, err := d.db.Exec("DELETE FROM configs WHERE key = ?", key)
	return err
}

func readConfigOptions(rows *sql.Rows) ([]ConfigOption, error) {
	defer rows.Close()
	var options []ConfigOption
	for rows.Next() {
		var option ConfigOption
		err := rows.Scan(&option.Key, &option.Value)
		if err != nil {
			return nil, err
		}
		options = append(options, option)
	}
	err := rows.Err()
	if err != nil {
		return nil, err

	}

	return options, nil
}

func (d *ConfigsDao) GetConfigOptions() ([]ConfigOption, error) {
	rows, err := d.db.Query("SELECT key, value FROM configs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readConfigOptions(rows)
}
