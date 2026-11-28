package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultSMTPFrom            = "no-reply@soc-trainer.local"
	DefaultSMTPPort            = 587
	DefaultEmailSendTimeout    = 15 * time.Second
	VerificationTokenTTL       = 24 * time.Hour
	VerificationResendCooldown = 60 * time.Second
	TLSModeSTARTTLS            = "starttls"
	TLSModeImplicit            = "tls"
	TLSModeNone                = "none"
)

// RequireEmailVerification returns whether new registrations must verify email ownership
// before login is permitted. Defaults to true when unset or invalid.
func RequireEmailVerification() bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("REQUIRE_EMAIL_VERIFICATION")))
	switch raw {
	case "false", "0", "no", "off":
		return false
	default:
		return true
	}
}

// ResolveAppBaseURL returns the public frontend origin used to build verification links.
func ResolveAppBaseURL() string {
	for _, explicit := range []string{os.Getenv("APP_BASE_URL"), os.Getenv("FRONTEND_URL")} {
		candidate := strings.TrimRight(strings.TrimSpace(explicit), "/")
		if candidate != "" && (strings.HasPrefix(candidate, "http://") || strings.HasPrefix(candidate, "https://")) {
			return candidate
		}
	}
	return "http://localhost"
}

// SMTPConfig holds outbound SMTP transport settings resolved from environment variables.
type SMTPConfig struct {
	Host               string
	Port               int
	User               string
	Password           string
	From               string
	TLSMode            string // TLSModeSTARTTLS (default), TLSModeImplicit, or TLSModeNone
	insecureSkipVerify bool
}

// ResolveSMTPConfig loads SMTP configuration with safe bounds and defaults.
func ResolveSMTPConfig() SMTPConfig {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := ResolveBoundedIntEnv("SMTP_PORT", DefaultSMTPPort, 1, 65535)
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	password := os.Getenv("SMTP_PASSWORD")
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	if from == "" {
		if user != "" {
			from = user
		} else {
			from = DefaultSMTPFrom
		}
	}
	tlsMode := strings.ToLower(strings.TrimSpace(os.Getenv("SMTP_TLS_MODE")))
	switch tlsMode {
	case TLSModeSTARTTLS, TLSModeImplicit, TLSModeNone:
	default:
		tlsMode = TLSModeSTARTTLS
	}
	return SMTPConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		From:     from,
		TLSMode:  tlsMode,
	}
}

// GenerateVerificationToken returns a fresh CSPRNG URL-safe token, its lowercase SHA-256 hex digest,
// and its 24-hour expiration timestamp.
func GenerateVerificationToken() (rawToken, tokenHash string, expiresAt time.Time, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate verification token: %w", err)
	}
	rawToken = base64.RawURLEncoding.EncodeToString(buf)
	tokenHash = HashVerificationToken(rawToken)
	expiresAt = time.Now().UTC().Add(VerificationTokenTTL)
	return rawToken, tokenHash, expiresAt, nil
}

// HashVerificationToken returns the lowercase SHA-256 hex digest for a raw verification token.
func HashVerificationToken(rawToken string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(rawToken)))
	return hex.EncodeToString(sum[:])
}

// MailSender abstracts transactional verification email dispatch cleanly for production SMTP and tests.
type MailSender interface {
	SendVerificationEmail(ctx context.Context, recipientEmail, nickname, verificationURL string) error
}

// SMTPMailSender dispatches verification emails over standard net/smtp.
type SMTPMailSender struct {
	cfg SMTPConfig
}

// NewSMTPMailSenderFromEnv builds an SMTPMailSender configured from environment variables.
func NewSMTPMailSenderFromEnv() *SMTPMailSender {
	return NewSMTPMailSenderWithConfig(ResolveSMTPConfig())
}

func EffectiveMailSender(sender MailSender) MailSender {
	if sender != nil {
		return sender
	}
	return NewSMTPMailSenderFromEnv()
}

// NewSMTPMailSenderWithConfig creates an SMTPMailSender using an explicit SMTPConfig.
func NewSMTPMailSenderWithConfig(cfg SMTPConfig) *SMTPMailSender {
	if strings.TrimSpace(cfg.From) == "" {
		cfg.From = DefaultSMTPFrom
	}
	if cfg.Port <= 0 {
		cfg.Port = DefaultSMTPPort
	}
	if cfg.TLSMode == "" {
		cfg.TLSMode = TLSModeSTARTTLS
	}
	return &SMTPMailSender{cfg: cfg}
}

func sanitizeHeaderValue(v string) string {
	trimmed := strings.TrimSpace(v)
	if idx := strings.IndexAny(trimmed, "\r\n"); idx >= 0 {
		trimmed = trimmed[:idx]
	}
	return strings.TrimSpace(trimmed)
}

func extractEmailAddress(headerOrAddr string) string {
	cleaned := sanitizeHeaderValue(headerOrAddr)
	if parsed, err := mail.ParseAddress(cleaned); err == nil && parsed != nil && parsed.Address != "" {
		return parsed.Address
	}
	return cleaned
}

// BuildVerificationEmail constructs a dual-language (TH + EN) multipart/alternative MIME message.
func BuildVerificationEmail(fromAddr, recipientEmail, nickname, verificationURL string) (string, []byte) {
	subject := "[SOC Trainer] ยืนยันอีเมลเพื่อเปิดใช้งานบัญชี / Verify Your Email Address"
	cleanFrom := sanitizeHeaderValue(fromAddr)
	cleanTo := sanitizeHeaderValue(recipientEmail)
	cleanNickname := sanitizeHeaderValue(nickname)
	if cleanNickname == "" {
		cleanNickname = "Analyst"
	}
	cleanURL := sanitizeHeaderValue(verificationURL)

	encodedSubject := mime.BEncoding.Encode("UTF-8", subject)
	escapedNickname := html.EscapeString(cleanNickname)
	escapedURL := html.EscapeString(cleanURL)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	body.WriteString(fmt.Sprintf("From: %s\r\n", cleanFrom))
	body.WriteString(fmt.Sprintf("To: %s\r\n", cleanTo))
	body.WriteString(fmt.Sprintf("Subject: %s\r\n", encodedSubject))
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%q\r\n\r\n", mw.Boundary()))

	plainText := fmt.Sprintf(
		"สวัสดีคุณ %s,\n\n"+
			"ขอบคุณที่สมัครใช้งานแพลตฟอร์ม SOC Trainer กรุณายืนยันอีเมลของคุณภายใน 24 ชั่วโมง เพื่อเปิดใช้งานบัญชีและเข้าสู่ระบบ:\n%s\n\n"+
			"---\n\n"+
			"Hello %s,\n\n"+
			"Thank you for registering on SOC Trainer. Please verify your email address within 24 hours to activate your account and sign in:\n%s\n\n"+
			"If you did not create this account, please ignore this email.\n",
		cleanNickname, cleanURL, cleanNickname, cleanURL,
	)

	textPart, _ := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/plain; charset=UTF-8"},
		"Content-Transfer-Encoding": {"quoted-printable"},
	})
	qpwText := quotedprintable.NewWriter(textPart)
	_, _ = qpwText.Write([]byte(strings.ReplaceAll(plainText, "\n", "\r\n")))
	_ = qpwText.Close()

	htmlText := fmt.Sprintf(
		`<!DOCTYPE html>
<html lang="th">
<body style="margin:0;padding:24px;background:#0f172a;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;color:#f8fafc;">
  <div style="max-width:600px;margin:0 auto;background:#111827;border:1px solid #334155;border-radius:10px;padding:28px;">
    <h2 style="margin:0 0 12px 0;font-size:20px;color:#ffffff;">SOC Trainer — ยืนยันอีเมล / Verify Email</h2>
    <p style="margin:0 0 12px 0;line-height:1.6;color:#cbd5e1;">
      สวัสดีคุณ <strong>%s</strong> กรุณากดปุ่มด้านล่างเพื่อยืนยันอีเมลและเปิดใช้งานบัญชีของคุณ (ลิงก์หมดอายุภายใน <strong>24 ชั่วโมง</strong>)
    </p>
    <p style="margin:0 0 20px 0;line-height:1.6;color:#94a3b8;">
      Hello <strong>%s</strong>, please click the button below to confirm your email ownership and activate login access (expires in <strong>24 hours</strong>).
    </p>
    <div style="margin:24px 0;text-align:center;">
      <a href="%s" style="display:inline-block;padding:12px 24px;background:#ffffff;color:#0f172a;text-decoration:none;border-radius:6px;font-weight:700;">
        ยืนยันอีเมล / Verify Email
      </a>
    </div>
    <p style="margin:16px 0 6px 0;font-size:13px;color:#94a3b8;">หากปุ่มใช้งานไม่ได้ สามารถคัดลอกลิงก์ด้านล่างไปวางในเบราว์เซอร์ได้ / Or paste this link into your browser:</p>
    <p style="margin:0;word-break:break-all;font-size:12px;color:#cbd5e1;">%s</p>
  </div>
</body>
</html>`,
		escapedNickname, escapedNickname, escapedURL, escapedURL,
	)

	htmlPart, _ := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/html; charset=UTF-8"},
		"Content-Transfer-Encoding": {"quoted-printable"},
	})
	qpwHTML := quotedprintable.NewWriter(htmlPart)
	_, _ = qpwHTML.Write([]byte(strings.ReplaceAll(htmlText, "\n", "\r\n")))
	_ = qpwHTML.Close()
	_ = mw.Close()

	return subject, body.Bytes()
}

// SendVerificationEmail connects to the configured SMTP endpoint (or skips safely when unconfigured)
// and transmits the dual-language verification message.
func (s *SMTPMailSender) SendVerificationEmail(ctx context.Context, recipientEmail, nickname, verificationURL string) error {
	if strings.TrimSpace(s.cfg.Host) == "" {
		log.Printf("[EMAIL] SMTP_HOST is not configured; skipping verification email delivery for recipient domain")
		return nil
	}

	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	dialer := &net.Dialer{Timeout: DefaultEmailSendTimeout}

	var conn net.Conn
	var err error
	if s.cfg.TLSMode == TLSModeImplicit {
		tlsDialer := &tls.Dialer{
			NetDialer: dialer,
			Config: &tls.Config{
				ServerName:         s.cfg.Host,
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: s.cfg.insecureSkipVerify,
			},
		}
		conn, err = tlsDialer.DialContext(ctx, "tcp", addr)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("smtp dial %s failed: %w", addr, err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(DefaultEmailSendTimeout))
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client init failed: %w", err)
	}
	defer client.Close()

	if s.cfg.TLSMode == TLSModeSTARTTLS {
		ok, _ := client.Extension("STARTTLS")
		if !ok {
			return fmt.Errorf("smtp server %s does not advertise STARTTLS (set SMTP_TLS_MODE=none for plaintext dev)", addr)
		}
		if err := client.StartTLS(&tls.Config{
			ServerName:         s.cfg.Host,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: s.cfg.insecureSkipVerify,
		}); err != nil {
			return fmt.Errorf("smtp STARTTLS failed: %w", err)
		}
	}

	if s.cfg.User != "" {
		auth := smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
	}

	fromEnvelope := extractEmailAddress(s.cfg.From)
	toEnvelope := extractEmailAddress(recipientEmail)
	if err := client.Mail(fromEnvelope); err != nil {
		return fmt.Errorf("smtp MAIL FROM failed: %w", err)
	}
	if err := client.Rcpt(toEnvelope); err != nil {
		return fmt.Errorf("smtp RCPT TO failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA open failed: %w", err)
	}
	_, mimeBytes := BuildVerificationEmail(s.cfg.From, toEnvelope, nickname, verificationURL)
	if _, err := w.Write(mimeBytes); err != nil {
		_ = w.Close()
		return fmt.Errorf("smtp DATA write failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp DATA close failed: %w", err)
	}

	_ = client.Quit()
	return nil
}

// BuildVerificationURL formats the frontend verification link from APP_BASE_URL cleanly.
func BuildVerificationURL(rawToken string) string {
	return ResolveAppBaseURL() + "/verify-email?token=" + url.QueryEscape(strings.TrimSpace(rawToken))
}

// DispatchVerificationEmailAsync sends the dual-language verification message on a background
// goroutine with a bounded timeout so registration/resend HTTP responses remain fast.
func DispatchVerificationEmailAsync(sender MailSender, recipientEmail, nickname, rawToken string) {
	if sender == nil {
		sender = NewSMTPMailSenderFromEnv()
	}
	verifyURL := BuildVerificationURL(rawToken)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), DefaultEmailSendTimeout)
		defer cancel()
		if err := sender.SendVerificationEmail(ctx, recipientEmail, nickname, verifyURL); err != nil {
			log.Printf("[EMAIL] Async verification delivery error: %v", err)
		}
	}()
}
