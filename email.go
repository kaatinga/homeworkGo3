package main

import (
	"fmt"
	"log"
	"net/smtp"
)

func (O *Order) SendEmail(s *Shop) error {

	// user we are authorizing as
	from := "noreplay@gmail.com"
	password := "123"

	// server we are authorized to send email through
	host := "smtp.gmail.com"

	// Create the authentication for the SendMail()
	// using PlainText, but other authentication methods are encouraged
	auth := smtp.PlainAuth("", from, password, host)

	// NOTE: Using the backtick here ` works like a heredoc, which is why all the
	// rest of the lines are forced to the beginning of the line, otherwise the
	// formatting is wrong for the RFC 822 style
	message := fmt.Sprintf(`To: "Test User" <%s>
From: "Kaatinga Test Shop" <%s>
Subject: Test Shop Notification
MIME-version: 1.0
Content-Type: text/plain; charset="UTF-8"

Hello, %s!

Thank you very much for your order.

The information about the order is below:
`, O.Email, from, O.Name)

	var total uint64

	for key, value := range O.Basket.list {

		good, ok := s.GetGood(key)
		if ok {
			message = fmt.Sprintf(`%sТовар: %s, Кол-во:%d, Цена: %d  
`, message, good.name, value, uint64(value)*good.price)

			total = total + uint64(value)*good.price
		}
	}

	message = fmt.Sprint(message, `
ИТОГО:`, total)

	fmt.Println(message)

	if err := smtp.SendMail(host+":587", auth, from, []string{O.Email}, []byte(message)); err != nil {
		log.Println("Error SendMail: ", err)
		return err
	}

	log.Println("Email Sent!")
	return nil
}
