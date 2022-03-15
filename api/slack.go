package api

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ThalesGroup/besec/api/models"
)

// SlackMessage captures the info needed for SlackSender to send and record a message
type SlackMessage struct {
	msg string
	uid string
}

// RequestAccessAlert sends a notification about a new user attempting to log in and on success records it in the user's local data
func RequestAccessAlert(rt *Runtime, user *models.User) {
	if rt.RequestAccessAlerts {
		sendAlert("New user request", rt, user)
	} else {
		log.Debug("New user request, alerts not configured")
	}
}

// NewUserAlert sends a notification about a user's first login and on success records it in the user's local data
func NewUserAlert(rt *Runtime, user *models.User) {
	if rt.NewUserAlerts {
		sendAlert("First sign-in from this authorized user", rt, user)
	} else {
		log.Debug("New user sign-in, alerts not configured")
	}
}

func sendAlert(event string, rt *Runtime, user *models.User) {
	// see https://api.slack.com/tools/block-kit-builder for the format

	tmpl := template.Must(template.New("msg").Parse(`{
		"blocks": [
			{
				"type": "section",
				"text": {
					"type": "mrkdwn",
					"text": "{{.Event}}:\n*{{.User.Name}}*"
				}
			},
			{
				"type": "section",
				"fields": [
					{
						"type": "mrkdwn",
						"text": "*Email:*\n{{.User.Email}}"
					},
					{
						"type": "mrkdwn",
						"text": "*Authenticated By:*\n{{.User.Provider}}"
					},
					{
						"type": "mrkdwn",
						"text": "*UID:*\n{{.User.UID}}"
					}
                ]{{if .User.PictureURL}},
                "accessory": {
                    "type": "image",
                    "image_url": "{{.User.PictureURL}}",
                    "alt_text": "user image"
                }
                {{- end}}
			}
		]
    }`))

	type data = struct {
		Event string
		User  *models.User
	}
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data{event, user})
	if err != nil {
		log.Errorf("Error formatting Slack message: %v", err)
		return
	}

	rt.SlackChan <- SlackMessage{msg: buf.String(), uid: user.UID}
	// the successful sending of the alert will be recorded by the channel receiver
}

// SlackSender sends messages received on c to Slack. Any duplicated messages that appear within a short period are not sent.
// If the message is successfully sent, it's recorded in the user's local record
func SlackSender(c chan SlackMessage, rt *Runtime, webhook string) {
	type sentMsg struct {
		msg  string
		time time.Time
	}

	if webhook == "" {
		log.Warn("No webhook configured, slack alerts won't be sent")
		return
	}

	expiry, _ := time.ParseDuration("1m")

	sent := []sentMsg{}
	for sm := range c {
		// Check for duplicates and filter out any expired messages, courtesy of https://github.com/golang/go/wiki/SliceTricks#filter-in-place
		n := 0
		duplicate := false
		for _, s := range sent {
			if time.Since(s.time) < expiry {
				sent[n] = s
				n++
				if s.msg == sm.msg {
					duplicate = true
				}
			}
		}
		sent = sent[:n]

		if !duplicate {
			err := sendMessage(rt, webhook, sm)
			if err == nil {
				sent = append(sent, sentMsg{msg: sm.msg, time: time.Now()})
			}
		} else {
			log.Debug("Got duplicate slack message, ignoring")
		}
	}
}

func sendMessage(rt *Runtime, webhook string, sm SlackMessage) error {
	log.Debug("Sending alert for user ", sm.uid)
	resp, err := http.Post(webhook, "application/json", strings.NewReader(sm.msg)) //nolint: gosec // this is a variable URL, but it's only configurable at deployment by admins
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("alerts.NewUser: Failed to send Slack notification message")
		return fmt.Errorf("failed to post to slack")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.WithFields(log.Fields{"user": sm.uid, "response code": resp.StatusCode, "response": body}).Warn("slack notification failed")
		return fmt.Errorf("got error response from slack")
	}

	err = rt.Store.UserCreationAlertSent(context.Background(), sm.uid)
	if err != nil {
		log.WithFields(log.Fields{"user": sm.uid, "error": err}).Warn("slack NewUserNotification: failed to record successful alert")
		// don't return an error - that will just lead to more duplicate messages
	}
	return nil
}
