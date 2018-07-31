package common

import (
	"github.com/recallsong/sogw/store"
	log "github.com/sirupsen/logrus"
)

var Store store.Store

func InitStore() store.Store {
	store, err := store.New(Config.Store.Url, Config.Store.Options)
	if err != nil {
		log.Fatal("[common] store.New failed : ", err)
		return nil
	}
	Store = store
	return store
}
