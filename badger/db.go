package badger

import (
	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/yoshino-s/go-framework/application"
	"github.com/yoshino-s/go-framework/configuration"
	"github.com/yoshino-s/go-framework/utils"
)

var _ application.Application = (*BadgerApp)(nil)

type BadgerApp struct {
	*application.EmptyApplication
	*badger.DB
	config config
}

func New() *BadgerApp {
	return &BadgerApp{
		EmptyApplication: application.NewEmptyApplication("Badger"),
	}
}

func (db *BadgerApp) Configuration() configuration.Configuration {
	return &db.config
}

func (db *BadgerApp) BeforeSetup(context.Context) {
	db.DB = utils.Must(badger.Open(badger.DefaultOptions(db.config.Path).WithLogger(&logger{db.Logger.Sugar()})))
}

func (db *BadgerApp) Close(context.Context) {
	utils.MustNoError(db.DB.Close())
}
