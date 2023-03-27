package util

import (
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	gomail "gopkg.in/gomail.v2"
)

func SendVerifyCode(VerifyCode, mail string) error {

	msg := gomail.NewMessage()
	msg.SetHeader("From", os.Getenv("SystemMail"))
	msg.SetHeader("To", mail)
	msg.SetHeader("Subject", "隧道肆月-註冊驗證碼")
	msg.SetBody("text/html", "<b>"+VerifyCode+"</b>")

	n := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("SystemMail"), os.Getenv("SystemMailPassword"))

	// Send the email
	if err := n.DialAndSend(msg); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
