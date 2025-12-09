package helpers

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(to, subject, body string) error {
    smtpHost := os.Getenv("SMTP_HOST")
    smtpPort := os.Getenv("SMTP_PORT")
    smtpUser := os.Getenv("SMTP_USER")
    smtpPass := os.Getenv("SMTP_PASS")

    msg := []byte(
        "Subject: " + subject + "\n" +
            "MIME-Version: 1.0;\n" +
            "Content-Type: text/html; charset=\"UTF-8\";\n\n" +
            body,
    )

    auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
    return smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{to}, msg)
}

func InviteEmailHTML(url string) string {
    return fmt.Sprintf(`
        <h2>You are invited!</h2>
        <p>Click below to complete your account setup:</p>
        <a href="%s" style="padding:10px 20px;background:#007bff;color:#fff;text-decoration:none;border-radius:6px;">Accept Invite</a>
        <p>If you did not expect this email, ignore it.</p>
    `, url)
}
