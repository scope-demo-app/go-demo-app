package main

import (
	"fmt"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/gofrs/uuid"
)

type (
	Rating struct {
		Id uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
		RestaurantId uuid.UUID `sql:",type:uuid"`
		Name string
		Description string
		Value int8
	}
)

var (
	db *pg.DB
)

func initConnection() *pg.DB {
	if db == nil {
		db = pg.Connect(&pg.Options{
			User:     "restUser",
			Password: "restPassw0rd",
			Database: "restaurant_ratings",
			Addr:     "localhost:5432",
		})
	}
	return db
}


func createSchema() error {
	db := initConnection()
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	for _, model := range []interface{}{(*Rating)(nil)} {
		err := db.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}

		var tableName string
		_, err = db.Model(model).Query(&tableName, "SELECT '?TableName'")
		if err != nil {
			panic(err)
		}
		tableName = tableName[1 : len(tableName)-1]
		db.Exec(fmt.Sprintf("CREATE INDEX %s_by_restaurant_id ON %s (restaurant_id)", tableName, tableName))
	}
	return nil
}
