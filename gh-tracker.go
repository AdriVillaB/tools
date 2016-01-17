package main

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)


// Configuration struct that holds the tracker runtime configuration
type Configuration struct {
	Environment string              `toml:"environment"`
	API         apiConfiguration    `toml:"api"`
	Mail        mailConfiguration   `toml:"mail"`
	Logger      loggerConfiguration `toml:"logger"`
}

type apiConfiguration struct {
	TokenAccess string `toml:"token_access"`
}

type mailConfiguration struct {
	Server      string `toml:"server"`
	SenderMail  string `toml:"sender_mail"`
	ReceiptMail string `toml:"receipt_mail"`
}

type loggerConfiguration struct {
	File   string `toml:"prefix"`
	Prefix string `toml:"file"`
}

// Tracker is the struct that holds the configuration for the tracker client
type Tracker struct {
	config Configuration
	logger *log.Logger
}

func (t Tracker) loadConfigFromFile(configFile string, config *Configuration) error {
	_, err := os.Stat(configFile)
	if err != nil {
		log.Fatal("Config file is missing: ", configFile)
	}
	
	//config := Configuration{}
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatal(err)
	}
		
	return nil
}

func (t Tracker) sendMail(body string) error {
	// Connect to the remote SMTP server.
	c, err := smtp.Dial(t.config.Mail.ReceiptMail)
	if err != nil {
		return err
	}
	defer c.Close()

	// Set the sender and recipient.
	c.Mail(t.config.Mail.SenderMail)
	c.Rcpt(t.config.Mail.ReceiptMail)

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()

	buf := bytes.NewBufferString(body)
	if _, err = buf.WriteTo(wc); err != nil {
		return err
	}
	return nil
}

func initialize() (Tracker, error) {
	tracker := Tracker{}

	//Load configuration from file
	err := tracker.loadConfigFromFile("config.toml", &tracker.config)
	if err != nil {
		return tracker, err
	}

	// Configure logger
	var loggerOut *os.File
	if tracker.config.Environment == "development" {
		loggerOut = os.Stderr
	} else if tracker.config.Environment == "production" {
		loggerOut, err = os.Open(tracker.config.Logger.File)
		if err != nil {
			panic("Cannot open log file")
		}
	} else {
		loggerOut = os.Stderr
	}
	tracker.logger = log.New(loggerOut, tracker.config.Logger.Prefix, log.Ldate|log.Ltime|log.Llongfile)
	return tracker, nil
}

func main() {
	tracker, err := initialize()
	if err != nil {
		tracker.logger.Panicf("Error initializing the tracker: %s", err.Error())
	}
	
	tokenService := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tracker.config.API.TokenAccess},
	)
	tokenContext := oauth2.NewClient(oauth2.NoContext, tokenService)

	client := github.NewClient(tokenContext)

	// Get unread notifications
	//opt := &github.NotificationListOptions{All: true}
	notifications, _, err := client.Activity.ListNotifications(nil)
	if err != nil {
		tracker.logger.Printf("Error listing notifications: %s", err.Error())
	}
	//fmt.Printf("Lenght of notifications array: %d\n", len(notifications))

	//If there are some notification, send mail
	if len(notifications) > 0 {
		bodyStr := fmt.Sprintf("There are %d new nofications in your GitHub account.\n\n", len(notifications))

		for _, notification := range notifications {
			var readed string 
			if *notification.Unread {
				readed = "Not readed"
			} else{
				readed = "Readed"
			}
			bodyStr += fmt.Sprintf("\t%s\t[%s]\t%s: %s. \n", readed, *notification.ID,*notification.Repository.FullName, *notification.Subject.Title)
		}

		tracker.logger.Println(bodyStr)
		tracker.sendMail(bodyStr)
	}
}
