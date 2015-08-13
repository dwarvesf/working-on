package main

import (
	"os"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	log "github.com/Sirupsen/logrus"
	"github.com/carlescere/scheduler"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
)

const (
	dbName string = "working"
)

var dbURL = os.Getenv("MONGOLAB_URI")

func main() {

	// Read configuration from file and env
	port := os.Getenv("PORT")
	if dbURL == "" {
		log.Fatal("No connection string provided")
		os.Exit(1)
	}

	// Setup schedule jobs
	remindJob := remind
	digestJob := digest
	scheduler.Every().Day().At("23:00").Run(remindJob)
	scheduler.Every().Day().At("09:30").Run(digestJob)

	// Prepare router
	router := gin.New()
	router.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))
	router.POST("/d", addDoneItem)
	router.POST("/l", login)

	// Start server
	router.Run(":" + port)
}

func getDBConnection() *mgo.Database {

	session, err := mgo.Dial(dbURL)
	if err != nil {
		panic(err)
	}

	defer session.Clone()
	return session.DB(dbName)
}

// Message will be passed to server with '-' prefix via various way
//	+ Direct message with the bots
//	+ Use slash command `/working <message>`
//	+ Use cli `working-on <message>`
//	+ Might use Chrome plugin
//	+ ...
// Token is secondary param to indicate the user
func addDoneItem(ctx *gin.Context) {

	type Item struct {
		ID        bson.ObjectId `json:"id" bson:"_id"`
		UserID    string        `json:"user_id" bson:"user_id"`
		Name      string        `json:"user_name" bson:"user_name"`
		Text      string        `json:"text" bson:"text"`
		CreatedAt time.Time     `json:"created_at" bson:"created_at"`
	}

	// Parse token and message
	var item Item
	if ctx.Bind(&item) == nil {

		// Add Item to database
		session, err := mgo.Dial(dbURL)
		if err != nil {
			panic(err)
		}
		defer session.Close()

		c := session.DB(dbName).C("items")
		err = c.Insert(item)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

// Log user in and return access token
func login(c *gin.Context) {

	// Retrieve username and password

	// Try to get some info from Slack

	// Store access token for user

	// Return access token for later usage
}

// Post reminder to Slack channel
func remind() {

}

// Post summary to Slack channel
func digest() {

}
