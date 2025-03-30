package dao

import "database/sql"

type Peripheral struct {
	ID   int64
	Key  string
	Name string
}

type PeripheralDao struct {
	db *sql.DB
}

func NewPeripheralDao(db *sql.DB) (*PeripheralDao, error) {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS peripherals (
		key  string NOT NULL,
		name string NOT NULL
	);
	`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &PeripheralDao{db: db}, nil
}

func (d *PeripheralDao) Create(peripheral *Peripheral) error {
	sqlStmt := `INSERT INTO peripherals (key, name) VALUES (?, ?);`
	_, err := d.db.Exec(sqlStmt, peripheral.Key, peripheral.Name)

	return err
}

func (d *PeripheralDao) Search(search string) ([]Peripheral, error) {
	sqlStmt := `SELECT id, key, name FROM peripherals WHERE name LIKE ? LIMIT 100;`
	rows, err := d.db.Query(sqlStmt, "%"+search+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peripherals []Peripheral
	for rows.Next() {
		var peripheral Peripheral
		if err := rows.Scan(&peripheral.ID, &peripheral.Key, &peripheral.Name); err != nil {
			return nil, err
		}
		peripherals = append(peripherals, peripheral)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return peripherals, nil
}

func (d *PeripheralDao) Delete(key string) error {
	sqlStmt := `DELETE FROM peripherals WHERE key = ?;`
	_, err := d.db.Exec(sqlStmt, key)

	return err
}
