package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/mgo.v2/bson"

	log "github.com/Sirupsen/logrus"
	"github.com/bmuller/arrow/lib"
	"github.com/carlescere/scheduler"
	"github.com/dwarvesf/working-on/db"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/nlopes/slack"
)

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

	// Setup schedule jobs
	digestJob := postDigest
	scheduler.Every().Day().At("09:30").Run(digestJob)

	// Prepare router
	router := gin.New()
	router.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))
	router.POST("/on", addItem)

	// Start server
	router.Run(":" + port)

	// postDigest()
}

// Message will be passed to server with '-' prefix via various way
//	+ Direct message with the bots
//	+ Use slash command `/working <message>`
//	+ Use cli `working-on <message>`
//	+ Might use Chrome plugin
//	+ ...
// Token is secondary param to indicate the user
func addItem(c *gin.Context) {

	// Parse token and message
	var item Item

	text := c.PostForm("text")
	userID := c.PostForm("user_id")
	userName := c.PostForm("user_name")

	item.ID = bson.NewObjectId()
	item.CreatedAt = time.Now()
	item.Name = userName
	item.UserID = userID
	item.Text = text
	fmt.Printf("Message: %v\n", item)

	ctx, err := db.NewContext()
	if err != nil {
		panic(err)
	}

	defer ctx.Close()

	// Add Item to database
	err = ctx.C("items").Insert(item)
	if err != nil {
		log.Fatalln(err)
	}
}

// Post summary to Slack channel
func postDigest() {

	channel := "#support"
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

	ctx, err := db.NewContext()
	if err != nil {
		panic(err)
	}

	defer ctx.Close()

	log.Info("Preparing data")
	// If count > 0, it means there is data to show
	count := 0
	title := "Yesterday I learnt"
	params := slack.PostMessageParameters{}
	fields := []slack.AttachmentField{}

	log.Info(arrow.Yesterday())
	log.Info(arrow.Now())
	log.Info(time.Now())

	// Prepare attachment of done items
	for _, user := range users {

		if user.IsBot || user.Deleted {
			continue
		}

		log.Info("Process user: " + user.Name)

		// Query done items from Database
		var values string
		var items []Item

		err = ctx.C("items").Find(bson.M{
			"user_id":    user.Id,
			"created_at": bson.M{"$gt": arrow.Yesterday()},
		}).All(&items)

		if err != nil {
			log.Fatal("Cannot query done items.")
			os.Exit(1)
		}

		if len(items) > 0 {
			count = count + 1
		}

		log.Info(len(items))

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

	params.IconURL = "http://i.imgur.com/fLcxkel.png"
	params.Username = "oshin"

	if count > 0 {
		s.PostMessage(channel, title, params)
	}
}
