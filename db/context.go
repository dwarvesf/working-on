package db

import (
	"os"

	"gopkg.in/mgo.v2"
)

type Context struct {
	Database *mgo.Database
}

func NewContext() (*Context, error) {

	dbURL := os.Getenv("MONGOLAB_URI")
	AuthDatabase := os.Getenv("DB_NAME")
	session, err := mgo.Dial(dbURL)
	if err != nil {
		return nil, err
	}

	ctx := &Context{
		Database: session.Clone().DB(AuthDatabase),
	}

	return ctx, err
}

func (c *Context) Close() {
	c.Database.Session.Close()
}

//C is a convenience function to return a collection from the context database.
func (c *Context) C(name string) *mgo.Collection {
	return c.Database.C(name)
}
