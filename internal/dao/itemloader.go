package dao

import (
	"context"
)

type ItemDeferedLoader struct {
	items map[string]*Item
	dao   *ItemsDao
}

func (l *ItemDeferedLoader) ToContext(parent context.Context) (context.Context, error) {
	err := l.dao.FindItemsIndexed(l.items)
	if err != nil {
		return nil, err
	}
	return context.WithValue(parent, "items", l.items), nil
}

func (l *ItemDeferedLoader) AddUid(uid string) *ItemDeferedLoader {
	l.items[uid] = nil
	return l
}

func (l *ItemDeferedLoader) AddUids(uid []string) *ItemDeferedLoader {
	for _, i := range uid {
		l.AddUid(i)
	}
	return l
}

func (l *ItemDeferedLoader) FromRecipe(r *Recipe) *ItemDeferedLoader {
	for _, i := range r.Ingredients {
		l.AddUid(i.ItemUID)
	}
	for _, i := range r.Results {
		l.AddUid(i.ItemUID)
	}
	return l
}

func (l *ItemDeferedLoader) FromRecipes(r []*Recipe) *ItemDeferedLoader {
	for _, i := range r {
		l.FromRecipe(i)
	}
	return l
}
