package main

import (
	"time"

	"github.com/tanishqv/bnb-bookings/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

func listenForMail() {
	go func() {
		for {
			msg := <-app.MailChan
			sendMessage(msg)
		}
	}()
}

func sendMessage(mData models.MailData) {
	server := mail.NewSMTPClient()

	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		errorLog.Println(err)
	}

	email := mail.NewMSG()
	email.SetFrom(mData.From).AddTo(mData.To).SetSubject(mData.Subject)
	email.SetBody(mail.TextHTML, mData.Content)

	err = email.Send(client)
	if err != nil {
		errorLog.Println(err)
	} else {
		infoLog.Println("mail sent")
	}
}
