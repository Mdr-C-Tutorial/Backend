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
		// 添加调试信息
		fmt.Printf("SMTP Host: %s\n", config.AppConfig.Email.SMTPHost)
		fmt.Printf("SMTP Port: %d\n", config.AppConfig.Email.SMTPPort)
		fmt.Printf("SMTP User: %s\n", config.AppConfig.Email.SMTPUser)
		// 注意：不要打印密码
		fmt.Printf("Dialer Host: %s\n", dialer.Host)
		fmt.Printf("Dialer Port: %d\n", dialer.Port)
	})
}

func SendVerificationEmail(user models.User) error {
	fmt.Printf("DEBUG: user.ID: %s\n", user.ID.String())
	initDialer()
	token, err := GenerateEmailVerificationToken(user.ID.String())
	if err != nil {
		fmt.Printf("生成邮箱验证Token失败: %v\n", err) // 添加错误日志
		return err
	}

	verificationLink := fmt.Sprintf("http://localhost:8080/api/auth/verify-email?token=%s", token)

	m := gomail.NewMessage()
	// 使用更简单的格式
	m.SetHeader("From", config.AppConfig.Email.FromEmail)
	m.SetHeader("To", user.Email)
	m.SetBody("text/html", fmt.Sprintf(`
		<h2>您好 %s,</h2>
		<p>请点击以下链接验证您的邮箱：</p>
		<a href="%s">验证邮箱</a>
		<p>此链接将在10分钟后失效。</p>
	`, user.Username, verificationLink))

	err = dialer.DialAndSend(m)
	if err != nil {
		fmt.Printf("发送验证邮件失败: %v\n", err) // 添加错误日志
	}
	return err
}

func SendFeedbackEmail(to string, content string, userID string) error {
	initDialer()
	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	m := gomail.NewMessage()
	// 使用更简单的格式
	m.SetHeader("From", config.AppConfig.Email.FromEmail)
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
