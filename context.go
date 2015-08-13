package main

import (
	"os"

	"gopkg.in/mgo.v2"
)

type Context struct {
	Database *mgo.Database
}

func NewContext() (*Context, error) {

	// AuthDatabase := os.Getenv("DB_NAME")
	// AuthUserName := os.Getenv("DB_USER")
	// AuthPassword := os.Getenv("DB_PASS")
	// MongoDBHosts := os.Getenv("DB_HOST")

	// dialInfo := &mgo.DialInfo{
	// Addrs:    []string{MongoDBHosts},
	// Timeout:  60 * time.Second,
	// Database: AuthDatabase,
	// Username: AuthUserName,
	// Password: AuthPassword,
	// }

	// session, err := mgo.DialWithInfo(dialInfo)

	dbURL := os.Getenv("MONGOLAB_URI")
	AuthDatabase := os.Getenv("DB_NAME")
	session, err := mgo.Dial(dbURL)
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
