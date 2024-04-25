package app

import "github.com/asek-ll/aecc-server/internal/dao"

type App struct {
	Daos *dao.DaoProvider
}

func CreateApp() (*App, error) {
	daos, err := dao.NewDaoProvider()
	if err != nil {
		return nil, err
	}
	return &App{Daos: daos}, nil
}
