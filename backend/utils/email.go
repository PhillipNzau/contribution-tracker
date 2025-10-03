package utils

import (
	"log"
	"os"

	"gopkg.in/gomail.v2"
)

// SendEmail sends an email using Gmail SMTP with credentials from env variables
func SendEmail(to, subject, body string) error {
	// Get credentials from environment variables
	from := os.Getenv("EMAIL_FROM")       // e.g. your Gmail address
	password := os.Getenv("EMAIL_PASS")   // Gmail App Password (not your Gmail login password)

	if from == "" || password == "" {
		log.Println("missing EMAIL_FROM or EMAIL_PASS environment variables")
		return nil
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Gmail SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, from, password)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		log.Printf("failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("email successfully sent to %s", to)
	return nil
}
