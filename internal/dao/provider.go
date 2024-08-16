package dao

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DaoProvider struct {
	Clients          *ClientsDao
	Seqs             *SeqsDao
	Items            *ItemsDao
	Recipes          *RecipesDao
	Configs          *ConfigsDao
	Plans            *PlansDao
	Crafts           *CraftsDao
	ItemReserves     *ItemReserveDao
	ImporetedRecipes *ImportedRecipesDao
	RecipeTypes      *RecipeTypesDao
	Workers          *WorkersDao
	StoredTX         *StoredTXDao
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

	plansDao, err := NewPlansDao(db)
	if err != nil {
		return nil, err
	}

	craftsDao, err := NewCraftsDao(db)
	if err != nil {
		return nil, err
	}

	itemReserveDao, err := NewItemReserveDao(db)
	if err != nil {
		return nil, err
	}

	importedRecipesDao, err := NewImportedRecipesDao(db)
	if err != nil {
		return nil, err
	}

	recipeTypesDao, err := NewRecipeTypesDao(db)
	if err != nil {
		return nil, err
	}

	workersDao, err := NewWorkersDao(db)
	if err != nil {
		return nil, err
	}

	storedTXDao, err := NewStoredTXDao(db)
	if err != nil {
		return nil, err
	}

	return &DaoProvider{
		Clients:          clientsDao,
		Seqs:             seqsDao,
		Items:            itemsDao,
		Recipes:          recipesDao,
		Configs:          configsDao,
		Plans:            plansDao,
		Crafts:           craftsDao,
		ItemReserves:     itemReserveDao,
		ImporetedRecipes: importedRecipesDao,
		RecipeTypes:      recipeTypesDao,
		Workers:          workersDao,
		StoredTX:         storedTXDao,
	}, nil
}
