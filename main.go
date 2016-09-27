package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	log "github.com/Sirupsen/logrus"
	"github.com/bmuller/arrow/lib"
	"github.com/brandfolder/gin-gorelic"
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
	digestTime := os.Getenv("DIGEST_TIME")
	gorelic.InitNewrelicAgent(os.Getenv("NEW_RELIC_LICENSE_KEY"), "working", false)

	digestConfig, err := parseConfig("digest.json")
	if err != nil {
		log.Fatalln(err)
	}

	// Setup schedule digest jobs
	for _, i := range digestConfig.Items {
		// channel := os.Getenv("DIGEST_CHANNEL")
		// botToken := os.Getenv("BOT_TOKEN")

		var strictTag bool
		if i.Tags != nil {
			strictTag = true
		}
		digestJob := postDigest(i.Channel, i.Token, i.Tags, strictTag)
		_, err := scheduler.Every().Day().At(digestTime).Run(digestJob)
		if err != nil {
			log.Infoln(err)
		}
	}

	settingConfig, err := parseConfig("setting.json")
	if err != nil {
		log.Fatalln(err)
	}

	// Prepare router
	router := gin.New()
	router.Use(gorelic.Handler)
	router.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))

	// router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")
	router.POST("/on", on(*settingConfig))
	router.POST("/til", til(*settingConfig))

	// Start server
	router.Run(":" + port)
}

func til(config Configuration) func(c *gin.Context) {
	return func(c *gin.Context) {
		text := c.PostForm("text")
		text = strings.TrimSpace(text)

		if text == "" {
			log.Fatalln("Message is nil")
			return
		}

		userID := c.PostForm("user_id")
		userName := c.PostForm("user_name")

		text = text + " #til"
		addItem(text, userID, userName, config)
	}
}

func on(config Configuration) func(c *gin.Context) {
	return func(c *gin.Context) {
		text := c.PostForm("text")
		text = strings.TrimSpace(text)

		if text == "" {
			log.Fatalln("Message is nil")
			return
		}

		userID := c.PostForm("user_id")
		userName := c.PostForm("user_name")

		addItem(text, userID, userName, config)
	}
}

// Message will be passed to server with '-' prefix via various way
//	+ Direct message with the bots
//	+ Use slash command `/working <message>`
//	+ Use cli `working-on <message>`
//	+ Might use Chrome plugin
//	+ ...
// Token is secondary param to indicate the user
func addItem(text string, userID string, userName string, configuration Configuration) {

	// Parse token and message
	var item Item

	item.ID = bson.NewObjectId()
	item.CreatedAt = time.Now()
	item.Name = userName
	item.UserID = userID
	item.Text = text

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

	// Repost to the target channel
	channel := os.Getenv("WORKING_CHANNEL")
	botToken := os.Getenv("BOT_TOKEN")

	if botToken == "" {
		log.Fatal("No token provided")
	}

	// <@U024BE7LH|bob>
	title := "*" + userName + "* is working on: " + text

	postWorkingItem(botToken, channel, title)

	// Post item to project group
	for _, config := range configuration.Items {
		for _, tag := range config.Tags {
			if strings.Contains(text, tag) {
				log.Info("Hit" + tag)
				log.Info("Token: " + config.Token)

				// Post to target channel
				postWorkingItem(config.Token, config.Channel, title)
			}
		}
	}
}

func postWorkingItem(token string, channel string, text string) {
	s := slack.New(token)
	params := slack.PostMessageParameters{}
	params.IconURL = "http://i.imgur.com/fLcxkel.png"
	params.Username = "oshin"
	s.PostMessage(channel, text, params)
}

type Configuration struct {
	Items []ConfigurationItem `json:"items"`
}

type ConfigurationItem struct {
	Channel string   `json:"channel"`
	Tags    []string `json:"tags"`
	Token   string   `json:"token"`
}

func parseConfig(path string) (*Configuration, error) {
	var configuration Configuration

	// Parse configuration
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("Cannot read setting file")
	}

	err = json.Unmarshal(bytes, &configuration)
	if err != nil {
		return nil, errors.New("Cannot parse setting")
	}

	return &configuration, nil
}

// Post summary to Slack channel.
// Only post to specific channel when tags are met.
func postDigest(channel, botToken string, tags []string, strictTag bool) func() {
	return func() {
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
		title := " >> Yesterday I did: "
		params := slack.PostMessageParameters{}
		fields := []slack.AttachmentField{}

		yesterday := arrow.Yesterday().UTC()
		toDate := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC)

		// Prepare attachment of done items
		for _, user := range users {

			if user.IsBot || user.Deleted {
				continue
			}

			// log.Info("Process user: " + user.Name + " - " + user.Id)

			// Query done items from Database
			var values string
			var items []Item

			err = ctx.C("items").Find(
				bson.M{
					"$and": []bson.M{
						bson.M{"user_name": user.Name},
						bson.M{"created_at": bson.M{"$gt": toDate}},
					},
				}).All(&items)

			if err != nil {
				log.Fatal("Cannot query done items.")
				os.Exit(1)
			}

			for i, item := range items {
				// delete items which text does not contain tags
				if strictTag && tags != nil {
					var containTag bool
					for _, tag := range tags {
						if !strings.Contains(item.Text, tag) {
							continue
						}
						containTag = true
						break
					}

					// if item's Text doesn't contains any tags then don't
					// add it to the digest message (values) and remove it
					// from items slice.
					if !containTag {
						if len(items) == 1 {
							items = nil
						} else {
							items = append(items[:i], items[i+1:]...)
						}
						continue
					}
				}

				// construct text format
				values = "\t" + values + " + " + item.Text + "\n"
			}

			// <@U024BE7LH|bob>
			if len(items) > 0 {

				count = count + 1
				field := slack.AttachmentField{
					Title: "@" + user.Name,
					Value: values,
				}

				fields = append(fields, field)
			}
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
}
