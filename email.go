package main

import (
	"fmt"
	"log"
	"net/smtp"
)

func (O *Order) SendEmail(message string) error {
	// user we are authorizing as
	from := "noreplay@gmail.com"

	password := "123"

	// use we are sending email to
	to := "kaatinga@gmail.com"

	// server we are authorized to send email through
	host := "smtp.gmail.com"

	// Create the authentication for the SendMail()
	// using PlainText, but other authentication methods are encouraged
	auth := smtp.PlainAuth("", from, password, host)

	// NOTE: Using the backtick here ` works like a heredoc, which is why all the
	// rest of the lines are forced to the beginning of the line, otherwise the
	// formatting is wrong for the RFC 822 style
	message = fmt.Sprintf(`To: "Some User" <someuser@example.com>
From: "Other User" <otheruser@example.com>
Subject: Test Shop Notification

%encoder!
`, message)

	if err := smtp.SendMail(host+":587", auth, from, []string{to}, []byte(message)); err != nil {
		log.Println("Error SendMail: ", err)
		return err
	}

	fmt.Println("Email Sent!")
	return nil
}