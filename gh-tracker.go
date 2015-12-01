package main

import (
	"bytes"
	"fmt"
	"net/smtp"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// tokenAccess is the token of the GitHub API
const tokenAccess = "951024f5903220c3c8932756edead1f9d6898194"

func sendMail(body string) error {
	// Connect to the remote SMTP server.
	c, err := smtp.Dial("localhost:25")
	if err != nil {
		return err
	}
	defer c.Close()

	// Set the sender and recipient.
	c.Mail("gh-tracker@adrivillab.com")
	c.Rcpt("adri@adrivillabermudez.net")

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

func main() {
	tokenService := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tokenAccess},
	)
	tokenContext := oauth2.NewClient(oauth2.NoContext, tokenService)

	client := github.NewClient(tokenContext)

	// Get unread notifications
	//opt := &github.NotificationListOptions{All: true}
	notifications, _, err := client.Activity.ListNotifications(nil)
	if err != nil {
		panic("Error listing notifications")
	}
	//fmt.Printf("Lenght of notifications array: %d\n", len(notifications))

	//If there are some notification, send mail
	if len(notifications) > 0 {
		bodyStr := fmt.Sprintf("There are %d new nofications in your GitHub account.\n\n", len(notifications))

		for _, notification := range notifications {
			bodyStr += fmt.Sprintf("\t%s: %s. \t\tLeído: %t\n", *notification.ID, *notification.Subject.Title, *notification.Unread)
		}
		
		fmt.Println(bodyStr)
		sendMail(bodyStr)
	}
}
