package dao

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DaoProvider struct {
	Clients *ClientsDao
	Seqs    *SeqsDao
	Items   *ItemsDao
	Recipes *RecipesDao
	Configs *ConfigsDao
}

func NewDaoProvider() (*DaoProvider, error) {
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		return nil, err
	}

	clientsDao, err := NewClientsDao(db)
	if err != nil {
		return nil, err
	}

	seqsDao, err := NewSeqsDao(db)
	if err != nil {
		return nil, err
	}

	itemsDao, err := NewItemsDao(db)
	if err != nil {
		return nil, err
	}

	recipesDao, err := NewRecipesDao(db)
	if err != nil {
		return nil, err
	}

	configsDao, err := NewConfigsDao(db)
	if err != nil {
		return nil, err
	}

	return &DaoProvider{
		Clients: clientsDao,
		Seqs:    seqsDao,
		Items:   itemsDao,
		Recipes: recipesDao,
		Configs: configsDao,
	}, nil
}
