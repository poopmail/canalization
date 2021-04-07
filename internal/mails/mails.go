package mails

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/poopmail/canalization/internal/id"
	"github.com/poopmail/canalization/internal/shared"
	"github.com/sirupsen/logrus"
)

type mail struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Content content  `json:"content"`
}

type content struct {
	Plain string `json:"plain"`
	HTML  string `json:"html"`
}

// Receiver represents the task which receives and processes incoming mails
func Receiver(ctx context.Context, pubSub *redis.PubSub, mailboxes shared.MailboxService, messages shared.MessageService) {
	logrus.Info("Starting the mail receiving task")
	channel := pubSub.Channel()

loop:
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Shutting down the mail receiving task")
			return
		case msg := <-channel:
			// Decode the incoming mail
			decoded, err := base64.StdEncoding.DecodeString(msg.Payload)
			if err != nil {
				logrus.WithError(err).Error("error while decoding incoming mail")
				continue loop
			}

			// Unmarshal the incoming mail
			mail := new(mail)
			if err := json.Unmarshal(decoded, mail); err != nil {
				logrus.WithError(err).Error("error while unmarshalling incoming mail")
				continue loop
			}

			// Retrieve the corresponding mailboxes
			addresses := make([]string, 0, len(mail.To))
			for _, to := range mail.To {
				mailbox, err := mailboxes.Mailbox(to)
				if err != nil {
					logrus.WithError(err).Error()
					continue loop
				}
				if mailbox != nil {
					addresses = append(addresses, mailbox.Address)
				}
			}

			// Write the mail to the database
			now := time.Now().Unix()
			for _, address := range addresses {
				message := &shared.Message{
					ID:      id.Generate(),
					Mailbox: address,
					From:    mail.From,
					Subject: mail.Subject,
					Content: &shared.MessageContent{
						Plain: mail.Content.Plain,
						HTML:  mail.Content.HTML,
					},
					Created: now,
				}
				if err := messages.CreateOrReplace(message); err != nil {
					logrus.WithError(err).Error("error while creating message from incoming mail")
				}
			}
		}
	}
}
