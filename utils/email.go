package utils

import (
	"crypto/tls"
	"fmt"
	"sync"

	"gopkg.in/gomail.v2"

	"mdr/config"
	"mdr/database"
	"mdr/models"
)

var (
	dialer *gomail.Dialer
	once   sync.Once
)

func initDialer() {
	once.Do(func() {
		dialer = gomail.NewDialer(
			config.AppConfig.Email.SMTPHost,
			config.AppConfig.Email.SMTPPort,
			config.AppConfig.Email.SMTPUser,
			config.AppConfig.Email.SMTPPassword,
		)
		dialer.SSL = true
		dialer.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	})
}

func SendVerificationEmail(user models.User) error {
	initDialer()
	token, err := GenerateEmailVerificationToken(user.ID.String())
	if err != nil {
		return err
	}

	verificationLink := fmt.Sprintf("http://localhost:8080/api/auth/verify-email?token=%s", token)

	m := gomail.NewMessage()
	m.SetHeader("From", config.AppConfig.Email.FromEmail)
	m.SetHeader("To", user.Email)
	m.SetHeader("From", fmt.Sprintf("MCT验证 <%s>", config.AppConfig.Email.FromEmail))
	m.SetBody("text/html", fmt.Sprintf(`
		<h2>您好 %s,</h2>
		<p>请点击以下链接验证您的邮箱：</p>
		<a href="%s">验证邮箱</a>
		<p>此链接将在10分钟后失效。</p>
	`, user.Username, verificationLink))

	return dialer.DialAndSend(m)
}

func SendFeedbackEmail(to string, content string, userID string) error {
	initDialer()
	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	m := gomail.NewMessage()
	// 修改发件人格式，添加名称
	m.SetHeader("From", fmt.Sprintf("MCT反馈 <%s>", config.AppConfig.Email.FromEmail))
	m.SetHeader("To", to)
	m.SetHeader("Subject", "MCT FeedBack")
	body := fmt.Sprintf(`
用户反馈信息：
-------------------
用户ID: %s
用户名: %s
邮箱: %s
-------------------
反馈内容:
%s
`, user.ID, user.Username, user.Email, content)
	fmt.Printf("邮件内容:\n%s\n", body)
	m.SetBody("text/plain", body)

	err := dialer.DialAndSend(m)
	if err != nil {
		fmt.Printf("发送邮件错误: %v\n", err)
		return err
	}

	return nil
}
