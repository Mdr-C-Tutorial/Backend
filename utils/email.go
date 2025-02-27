package utils

import (
	"fmt"
	"net/smtp"

	"mdr/config"
	"mdr/models"
)

func SendVerificationEmail(user models.User) error {
	token, err := GenerateEmailVerificationToken(user.ID.String())
	if err != nil {
		return err
	}

	verificationLink := fmt.Sprintf("http://localhost:8080/api/auth/verify-email?token=%s", token)
	
	subject := "验证您的邮箱"
	body := fmt.Sprintf(`
		<h2>您好 %s,</h2>
		<p>请点击以下链接验证您的邮箱：</p>
		<a href="%s">验证邮箱</a>
		<p>此链接将在10分钟后失效。</p>
	`, user.Username, verificationLink)

	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", user.Email, subject, body)

	auth := smtp.PlainAuth("",
		config.AppConfig.Email.SMTPUser,
		config.AppConfig.Email.SMTPPassword,
		config.AppConfig.Email.SMTPHost,
	)

	return smtp.SendMail(
		fmt.Sprintf("%s:%d", config.AppConfig.Email.SMTPHost, config.AppConfig.Email.SMTPPort),
		auth,
		config.AppConfig.Email.FromEmail,
		[]string{user.Email},
		[]byte(msg),
	)
}