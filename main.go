package main

import (
	"os"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	log "github.com/Sirupsen/logrus"
	"github.com/bmuller/arrow/lib"
	"github.com/carlescere/scheduler"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/nlopes/slack"
)

const (
	dbName string = "working"
)

var dbURL = os.Getenv("MONGOLAB_URI")

type Item struct {
	ID        bson.ObjectId `json:"id" bson:"_id"`
	UserID    string        `json:"user_id" bson:"user_id"`
	Name      string        `json:"user_name" bson:"user_name"`
	Text      string        `json:"text" bson:"text"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
}

func main() {

	// Read configuration from file and env
	port := os.Getenv("PORT")
	if dbURL == "" {
		log.Fatal("No connection string provided")
		os.Exit(1)
	}

	// Setup schedule jobs
	digestJob := postDigest
	scheduler.Every().Day().At("09:30").Run(digestJob)

	// Prepare router
	router := gin.New()
	router.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))
	router.POST("/on", addItem)

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
func addItem(ctx *gin.Context) {

	// Parse token and message
	var item Item
	if ctx.Bind(&item) != nil {
		log.Fatal("Cannot parse data")
	}

	item.CreatedAt = time.Now()

	message := strings.TrimSpace(item.Text)
	if !strings.HasPrefix(message, "-") {
		panic("Wrong format: " + item.Text)
	}

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

// Post summary to Slack channel
func postDigest() {

	channel := "#general"
	botToken := os.Getenv("BOT_TOKEN")

	if botToken == "" {
		log.Fatal("No token provided")
		os.Exit(1)
	}

	s := slack.New(botToken)
	users, err := s.GetUsers()

	if err != nil {
		log.Fatal("Cannot get users")
		os.Exit(1)
	}

	itemCollection := getDBConnection().C("items")

	// If count > 0, it means there is data to show
	count := 0
	title := "Yesterday I learnt"
	params := slack.PostMessageParameters{}
	fields := []slack.AttachmentField{}

	// Prepare attachment of done items
	for _, user := range users {

		if user.IsBot || user.Deleted {
			continue
		}

		// Query done items from Database
		var values string
		var items []Item
		err = itemCollection.Find(bson.M{"user_id": user.Id, "created_at": bson.M{"$gt": arrow.Yesterday()}}).All(&items)
		if err != nil {
			log.Fatal("Cannot query done items.")
			os.Exit(1)
		}

		if len(items) > 0 {
			count = count + 1
		}

		for _, item := range items {
			values = values + item.Text + "\n"
		}

		field := slack.AttachmentField{
			Title: user.Name,
			Value: values,
		}

		fields = append(fields, field)
	}

	params.Attachments = []slack.Attachment{
		slack.Attachment{
			Color:  "#7CD197",
			Fields: fields,
		},
	}

	if count > 0 {
		s.PostMessage(channel, title, params)
	}
}
