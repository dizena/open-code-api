package mail

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"math/big"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	cfg config.SMTPConfig
}

func NewService(cfg config.SMTPConfig) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) IsEnabled() bool {
	return s.cfg.Enabled &&
		s.cfg.Host != "" &&
		s.cfg.Port > 0 &&
		s.cfg.User != "" &&
		s.cfg.Password != ""
}

func (s *Service) SendVerificationCode(to string, code string) error {
	if !s.IsEnabled() {
		log.Warnf("smtp: service not enabled, verification code for %s: %s", to, code)
		return nil
	}

	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("smtp: invalid recipient address %q: %w", to, err)
	}

	subject := "淡香雅小站的密钥"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f5f5f5;">
  <table role="presentation" width="100vw" cellpadding="0" cellspacing="0" style="background-color:#f5f5f5;padding:40px 0;">
    <tr>
      <td align="center">
        <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.1);">
          <tr>
            <td style="background:linear-gradient(135deg,#0ea5e9,#6366f1);padding:32px;text-align:center;">
              <h1 style="margin:0;color:#ffffff;font-size:24px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">淡香雅小站</h1>
              <p style="margin:8px 0 0;color:rgba(255,255,255,0.9);font-size:14px;">CodeApi平台</p>
            </td>
          </tr>
          <tr>
            <td style="padding:40px 32px;text-align:center;">
              <h2 style="margin:0 0 16px;color:#1f2937;font-size:20px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">您的验证码</h2>
              <p style="margin:0 0 24px;color:#6b7280;font-size:14px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">请使用以下 8 位数字验证码完成身份验证：</p>
              <div style="display:inline-block;background-color:#f0f9ff;border:2px solid #0ea5e9;border-radius:8px;padding:16px 32px;">
                <span style="font-size:32px;font-weight:700;color:#0ea5e9;letter-spacing:8px;font-family:'Courier New',monospace;">%s</span>
              </div>
              <p style="margin:24px 0 0;color:#9ca3af;font-size:12px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">验证码 10 分钟内有效，请勿泄露给他人。</p>
            </td>
          </tr>
          <tr>
            <td style="background-color:#f9fafb;padding:24px;text-align:center;border-top:1px solid #e5e7eb;">
              <p style="margin:0;color:#9ca3af;font-size:12px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">此邮件由系统自动发送，请勿直接回复。</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
	</table>
</body>
</html>
`, code)

	return s.sendEmail(to, subject, body)
}

// sendEmail sends email, using STARTTLS on port 587 or plain SMTP on other ports.
func (s *Service) sendEmail(to, subject, body string) error {
	message := s.buildMessage(to, subject, body)

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	client, err := smtp.Dial(addr)
	if err != nil {
		log.WithError(err).Error("smtp: dial failed")
		return fmt.Errorf("smtp: dial: %w", err)
	}
	defer client.Quit()

	// STARTTLS on port 587
	if s.cfg.Port == 587 {
		tlsConfig := &tls.Config{
			ServerName:         s.cfg.Host,
			InsecureSkipVerify: true,
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			log.WithError(err).Error("smtp: STARTTLS failed")
			return fmt.Errorf("smtp: starttls: %w", err)
		}
		log.Info("smtp: TLS handshake succeeded via STARTTLS")
	}

	auth := newInsecurePlainAuth(s.cfg.User, s.cfg.Password)
	if err = client.Auth(auth); err != nil {
		log.WithError(err).Error("smtp: auth failed")
		return fmt.Errorf("smtp: auth: %w", err)
	}

	if err = s.send(client, to, message); err != nil {
		return err
	}

	log.Infof("smtp: sent verification code to %s (port %d)", to, s.cfg.Port)
	return nil
}

// buildMessage constructs an SMTP message with headers and body.
func (s *Service) buildMessage(to, subject, body string) string {
	from := s.fromAddress()
	return fmt.Sprintf("To: %s\r\n", to) +
		fmt.Sprintf("From: %s\r\n", from) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body
}

// send performs the Mail/Rcpt/Data SMTP sequence.
func (s *Service) send(client *smtp.Client, to, message string) error {
	if err := client.Mail(s.fromAddress()); err != nil {
		log.WithError(err).Error("smtp: MAIL FROM failed")
		return fmt.Errorf("smtp: mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		log.WithError(err).Error("smtp: RCPT TO failed")
		return fmt.Errorf("smtp: rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		log.WithError(err).Error("smtp: DATA failed")
		return fmt.Errorf("smtp: data: %w", err)
	}
	defer w.Close()

	if _, err = w.Write([]byte(message)); err != nil {
		log.WithError(err).Error("smtp: write message failed")
		return fmt.Errorf("smtp: write: %w", err)
	}
	return nil
}

func (s *Service) fromAddress() string {
	return strings.TrimSpace(s.cfg.User)
}

type insecurePlainAuth struct {
	username string
	password string
}

func newInsecurePlainAuth(username, password string) smtp.Auth {
	return &insecurePlainAuth{
		username: username,
		password: password,
	}
}

func (a *insecurePlainAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	payload := "\x00" + a.username + "\x00" + a.password
	return "PLAIN", []byte(payload), nil
}

func (a *insecurePlainAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, fmt.Errorf("smtp: unexpected auth challenge: %s", string(fromServer))
	}
	return nil, nil
}

func GenerateCode() (string, error) {
	var b strings.Builder
	var s = "012356789"
	for i := 0; i < 8; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(s))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random code: %w", err)
		}
		b.WriteByte(s[n.Int64()])
	}
	
	return b.String(), nil
}
