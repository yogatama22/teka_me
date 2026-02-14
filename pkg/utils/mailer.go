package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"

	gomail "gopkg.in/gomail.v2"
)

func SendEmailGmail(to, subject, body string) error {
	log.Printf("📮 Preparing email to: %s", to) // TAMBAHKAN INI

	m := gomail.NewMessage()

	from := os.Getenv("SMTP_FROM")
	name := os.Getenv("SMTP_NAME")

	log.Printf("📧 SMTP Config - Host: %s, Port: %s, User: %s",
		os.Getenv("SMTP_HOST"),
		os.Getenv("SMTP_PORT"),
		os.Getenv("SMTP_USERNAME")) // TAMBAHKAN INI

	m.SetHeader("From", fmt.Sprintf("%s <%s>", name, from))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Printf("❌ Invalid SMTP_PORT: %v", err) // TAMBAHKAN INI
		return fmt.Errorf("invalid SMTP_PORT: %v", err)
	}

	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		port,
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
	)

	if os.Getenv("SMTP_SECURE") == "ssl" {
		d.SSL = true
		log.Println("🔒 Using SSL") // TAMBAHKAN INI
	} else {
		log.Println("🔓 Not using SSL") // TAMBAHKAN INI
	}

	log.Println("🔌 Connecting to SMTP server...") // TAMBAHKAN INI
	if err := d.DialAndSend(m); err != nil {
		log.Printf("❌ SMTP Error: %v", err) // TAMBAHKAN INI
		return err
	}

	log.Println("✅ SMTP connection successful") // TAMBAHKAN INI
	return nil
}
