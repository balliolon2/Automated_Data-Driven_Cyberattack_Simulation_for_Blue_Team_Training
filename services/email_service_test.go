package services

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/big"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRequireEmailVerificationEnv(t *testing.T) {
	cases := []struct {
		raw  string
		want bool
	}{
		{"", true},
		{"   ", true},
		{"true", true},
		{"TRUE", true},
		{"1", true},
		{"yes", true},
		{"invalid", true},
		{"false", false},
		{"FALSE", false},
		{" 0 ", false},
		{"no", false},
		{"off", false},
	}

	for _, tc := range cases {
		t.Run("raw_"+tc.raw, func(t *testing.T) {
			t.Setenv("REQUIRE_EMAIL_VERIFICATION", tc.raw)
			if got := RequireEmailVerification(); got != tc.want {
				t.Fatalf("RequireEmailVerification(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestResolveAppBaseURL(t *testing.T) {
	t.Run("explicit_app_base_url_trimmed", func(t *testing.T) {
		t.Setenv("APP_BASE_URL", " https://soc.example.ac.th/ ")
		t.Setenv("FRONTEND_URL", "http://localhost:5173")
		if got := ResolveAppBaseURL(); got != "https://soc.example.ac.th" {
			t.Fatalf("got %q, want https://soc.example.ac.th", got)
		}
	})

	t.Run("fallback_to_frontend_url", func(t *testing.T) {
		t.Setenv("APP_BASE_URL", "")
		t.Setenv("FRONTEND_URL", "https://frontend.ku.ac.th/")
		if got := ResolveAppBaseURL(); got != "https://frontend.ku.ac.th" {
			t.Fatalf("got %q, want https://frontend.ku.ac.th", got)
		}
	})

	t.Run("default_localhost_when_empty", func(t *testing.T) {
		t.Setenv("APP_BASE_URL", "")
		t.Setenv("FRONTEND_URL", "")
		if got := ResolveAppBaseURL(); got != "http://localhost" {
			t.Fatalf("got %q, want http://localhost", got)
		}
	})
}

func TestResolveSMTPConfigBoundsAndDefaults(t *testing.T) {
	t.Setenv("SMTP_HOST", " mailpit ")
	t.Setenv("SMTP_PORT", "1025")
	t.Setenv("SMTP_USER", "mailer@soc.local")
	t.Setenv("SMTP_PASSWORD", "secret")
	t.Setenv("SMTP_FROM", "")
	t.Setenv("SMTP_TLS_MODE", "NONE")

	cfg := ResolveSMTPConfig()
	if cfg.Host != "mailpit" || cfg.Port != 1025 || cfg.From != "mailer@soc.local" || cfg.TLSMode != "none" {
		t.Fatalf("unexpected SMTP config: %+v", cfg)
	}

	t.Setenv("SMTP_PORT", "999999")
	t.Setenv("SMTP_TLS_MODE", "weird")
	t.Setenv("SMTP_USER", "")
	cfgFallback := ResolveSMTPConfig()
	if cfgFallback.Port != 65535 || cfgFallback.TLSMode != "starttls" || cfgFallback.From != DefaultSMTPFrom {
		t.Fatalf("unexpected fallback SMTP config: %+v", cfgFallback)
	}
}

func TestGenerateVerificationTokenAndHash(t *testing.T) {
	raw1, hash1, exp1, err := GenerateVerificationToken()
	if err != nil {
		t.Fatalf("GenerateVerificationToken failed: %v", err)
	}
	raw2, hash2, _, err := GenerateVerificationToken()
	if err != nil {
		t.Fatalf("second GenerateVerificationToken failed: %v", err)
	}
	if raw1 == "" || raw1 == raw2 || hash1 == hash2 {
		t.Fatalf("expected unique non-empty CSPRNG tokens and hashes")
	}
	if len(hash1) != 64 {
		t.Fatalf("expected 64-char SHA-256 hex digest, got %d", len(hash1))
	}
	expectedSum := sha256.Sum256([]byte(raw1))
	if hash1 != hex.EncodeToString(expectedSum[:]) || HashVerificationToken(raw1) != hash1 {
		t.Fatalf("token hash mismatch: got %s", hash1)
	}
	until := time.Until(exp1)
	if until < 23*time.Hour+59*time.Minute || until > 24*time.Hour+1*time.Minute {
		t.Fatalf("expected ~24h TTL, got %s", until)
	}
}

func decodeVerificationMIMEBodies(t *testing.T, rawMIME []byte) (mail.Header, string, string) {
	t.Helper()
	parsedMsg, err := mail.ReadMessage(bytes.NewReader(rawMIME))
	if err != nil {
		t.Fatalf("mail.ReadMessage failed: %v", err)
	}
	mediaType, params, err := mime.ParseMediaType(parsedMsg.Header.Get("Content-Type"))
	if err != nil || !strings.HasPrefix(mediaType, "multipart/alternative") {
		t.Fatalf("expected multipart/alternative Content-Type, got %q (err=%v)", parsedMsg.Header.Get("Content-Type"), err)
	}
	mr := multipart.NewReader(parsedMsg.Body, params["boundary"])
	var textBody, htmlBody string
	for {
		part, nextErr := mr.NextPart()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			t.Fatalf("multipart NextPart failed: %v", nextErr)
		}
		var reader io.Reader = part
		if strings.EqualFold(part.Header.Get("Content-Transfer-Encoding"), "quoted-printable") {
			reader = quotedprintable.NewReader(part)
		}
		decoded, readErr := io.ReadAll(reader)
		if readErr != nil {
			t.Fatalf("read MIME part failed: %v", readErr)
		}
		partCT := part.Header.Get("Content-Type")
		switch {
		case strings.Contains(partCT, "text/plain"):
			textBody = string(decoded)
		case strings.Contains(partCT, "text/html"):
			htmlBody = string(decoded)
		}
	}
	return parsedMsg.Header, textBody, htmlBody
}

func TestBuildVerificationEmailDualLanguageAndSanitization(t *testing.T) {
	rawToken := "test-token-123"
	url := "https://soc.example.com/verify-email?token=" + rawToken
	subj, rawMIME := BuildVerificationEmail(
		"SOC Trainer <noreply@soc.example.com>",
		"learner@soc.local\r\nBcc: attacker@evil.test",
		"<script>alert(1)</script>Alice",
		url,
	)

	if !strings.Contains(subj, "ยืนยันอีเมล") || !strings.Contains(subj, "Verify Your Email") {
		t.Fatalf("expected dual-language TH+EN subject, got %q", subj)
	}
	msg := string(rawMIME)
	if strings.Contains(msg, "Bcc: attacker@evil.test") {
		t.Fatalf("CRLF header injection was not stripped")
	}
	headers, textPart, htmlPart := decodeVerificationMIMEBodies(t, rawMIME)
	if gotTo := headers.Get("To"); gotTo != "learner@soc.local" {
		t.Fatalf("expected sanitized To header, got %q", gotTo)
	}
	if !strings.Contains(textPart, "ยืนยันอีเมล") || !strings.Contains(textPart, "verify your email address") || !strings.Contains(htmlPart, "ยืนยันอีเมล") || !strings.Contains(htmlPart, "Verify Email") {
		t.Fatalf("expected both Thai and English copy in decoded MIME parts")
	}
	if !strings.Contains(textPart, "24 ชั่วโมง") || !strings.Contains(textPart, "24 hours") || !strings.Contains(htmlPart, "24 ชั่วโมง") {
		t.Fatalf("expected 24-hour bilingual expiry notice")
	}
	if !strings.Contains(textPart, url) || !strings.Contains(htmlPart, url) {
		t.Fatalf("expected verification URL in decoded email parts")
	}
	if !strings.Contains(htmlPart, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("expected HTML escaping on nickname in decoded text/html part")
	}
	if strings.Contains(htmlPart, "<script>alert(1)</script>") {
		t.Fatalf("expected raw script tag absent from decoded HTML section")
	}
}

func TestSMTPMailSender_UnconfiguredLogsSafelyWithoutToken(t *testing.T) {
	var logBuf bytes.Buffer
	origOut := log.Writer()
	log.SetOutput(&logBuf)
	defer log.SetOutput(origOut)

	sender := NewSMTPMailSenderWithConfig(SMTPConfig{
		Host: "",
		Port: 587,
		From: DefaultSMTPFrom,
	})
	secretToken := "super-secret-verify-token-xyz"
	verifyURL := "http://localhost/verify-email?token=" + secretToken

	if err := sender.SendVerificationEmail(context.Background(), "student@example.com", "student", verifyURL); err != nil {
		t.Fatalf("expected nil error on unconfigured SMTP skip, got %v", err)
	}
	logged := logBuf.String()
	if !strings.Contains(logged, "SMTP_HOST") {
		t.Fatalf("expected safe operational log noticing missing SMTP_HOST, got %q", logged)
	}
	if strings.Contains(logged, secretToken) || strings.Contains(logged, verifyURL) {
		t.Fatalf("raw verification token or URL leaked into log output: %q", logged)
	}
}

func TestSMTPMailSender_LoopbackDeliveryPlainAndSTARTTLS(t *testing.T) {
	t.Run("plain_loopback_delivery", func(t *testing.T) {
		addr, captured, shutdown := startFakeSMTPServer(t, nil)
		defer shutdown()

		host, portStr, _ := net.SplitHostPort(addr)
		var port int
		fmt.Sscanf(portStr, "%d", &port)

		sender := NewSMTPMailSenderWithConfig(SMTPConfig{
			Host:    host,
			Port:    port,
			From:    "noreply@soc.local",
			TLSMode: "none",
		})
		verifyURL := "http://localhost/verify-email?token=abc123"
		if err := sender.SendVerificationEmail(context.Background(), "alice@soc.local", "alice", verifyURL); err != nil {
			t.Fatalf("SendVerificationEmail plain loopback failed: %v", err)
		}

		msg := captured()
		headers, textPart, htmlPart := decodeVerificationMIMEBodies(t, []byte(msg))
		if headers.Get("To") != "alice@soc.local" || !strings.Contains(textPart, verifyURL) || !strings.Contains(htmlPart, verifyURL) {
			t.Fatalf("expected captured SMTP payload to include recipient and verification link, got:\n%s", msg)
		}
	})

	t.Run("implicit_tls_loopback_delivery", func(t *testing.T) {
		cert := generateSelfSignedLoopbackCert(t)
		addr, captured, shutdown := startFakeSMTPServer(t, &tls.Config{
			Certificates: []tls.Certificate{cert},
		})
		defer shutdown()

		host, portStr, _ := net.SplitHostPort(addr)
		var port int
		fmt.Sscanf(portStr, "%d", &port)

		sender := NewSMTPMailSenderWithConfig(SMTPConfig{
			Host:               host,
			Port:               port,
			From:               "noreply@soc.local",
			TLSMode:            "tls",
			insecureSkipVerify: true,
		})
		verifyURL := "https://soc.example.com/verify-email?token=tls123"
		if err := sender.SendVerificationEmail(context.Background(), "bob@soc.local", "bob", verifyURL); err != nil {
			t.Fatalf("SendVerificationEmail TLS loopback failed: %v", err)
		}
		_, textPart, htmlPart := decodeVerificationMIMEBodies(t, []byte(captured()))
		if !strings.Contains(textPart, verifyURL) || !strings.Contains(htmlPart, verifyURL) {
			t.Fatalf("expected captured TLS SMTP payload to contain verification link")
		}
	})

	t.Run("starttls_upgrade_and_missing_starttls_rejection", func(t *testing.T) {
		cert := generateSelfSignedLoopbackCert(t)
		addrUpgrade, capturedUpgrade, shutdownUpgrade := startFakeSTARTTLSServer(t, &tls.Config{
			Certificates: []tls.Certificate{cert},
		})
		defer shutdownUpgrade()

		host, portStr, _ := net.SplitHostPort(addrUpgrade)
		var port int
		fmt.Sscanf(portStr, "%d", &port)

		senderOK := NewSMTPMailSenderWithConfig(SMTPConfig{
			Host:               host,
			Port:               port,
			From:               "noreply@soc.local",
			TLSMode:            "starttls",
			insecureSkipVerify: true,
		})
		verifyURL := "https://soc.example.com/verify-email?token=starttls123"
		if err := senderOK.SendVerificationEmail(context.Background(), "carol@soc.local", "carol", verifyURL); err != nil {
			t.Fatalf("expected STARTTLS upgrade delivery to succeed, got %v", err)
		}
		_, textPart, htmlPart := decodeVerificationMIMEBodies(t, []byte(capturedUpgrade()))
		if !strings.Contains(textPart, verifyURL) || !strings.Contains(htmlPart, verifyURL) {
			t.Fatalf("expected captured STARTTLS payload to contain verification link")
		}

		// Plain server without STARTTLS must fail cleanly instead of silently downgrading.
		addrPlain, _, shutdownPlain := startFakeSMTPServer(t, nil)
		defer shutdownPlain()
		hPlain, pPlainStr, _ := net.SplitHostPort(addrPlain)
		var pPlain int
		fmt.Sscanf(pPlainStr, "%d", &pPlain)
		senderStrict := NewSMTPMailSenderWithConfig(SMTPConfig{
			Host:    hPlain,
			Port:    pPlain,
			From:    "noreply@soc.local",
			TLSMode: "starttls",
		})
		if err := senderStrict.SendVerificationEmail(context.Background(), "dave@soc.local", "dave", verifyURL); err == nil || !strings.Contains(err.Error(), "STARTTLS") {
			t.Fatalf("expected explicit STARTTLS error when unadvertised, got %v", err)
		}
	})

	t.Run("context_cancellation_fails_promptly", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		sender := NewSMTPMailSenderWithConfig(SMTPConfig{
			Host:    "127.0.0.1",
			Port:    65432,
			From:    "noreply@soc.local",
			TLSMode: "none",
		})
		if err := sender.SendVerificationEmail(ctx, "eve@soc.local", "eve", "http://localhost/verify-email?token=x"); err == nil {
			t.Fatalf("expected error on canceled context")
		}
	})
}

func startFakeSTARTTLSServer(t *testing.T, tlsCfg *tls.Config) (string, func() string, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen for STARTTLS fake server: %v", err)
	}
	var mu sync.Mutex
	var dataBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			return
		}
		defer conn.Close()
		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		writeLine := func(line string) {
			_, _ = rw.WriteString(line + "\r\n")
			_ = rw.Flush()
		}
		writeLine("220 localhost ESMTP FakeSTARTTLS")
		inData := false
		for {
			line, readErr := rw.ReadString('\n')
			if readErr != nil {
				return
			}
			trimmed := strings.TrimRight(line, "\r\n")
			if inData {
				if trimmed == "." {
					inData = false
					writeLine("250 2.0.0 Ok")
					continue
				}
				mu.Lock()
				dataBuf.WriteString(line)
				mu.Unlock()
				continue
			}
			upper := strings.ToUpper(trimmed)
			switch {
			case strings.HasPrefix(upper, "EHLO"):
				writeLine("250-localhost")
				writeLine("250 STARTTLS")
			case upper == "STARTTLS":
				writeLine("220 Ready to start TLS")
				tlsConn := tls.Server(conn, tlsCfg)
				if err := tlsConn.Handshake(); err != nil {
					return
				}
				conn = tlsConn
				rw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
			case strings.HasPrefix(upper, "MAIL FROM:"), strings.HasPrefix(upper, "RCPT TO:"):
				writeLine("250 2.1.0 Ok")
			case upper == "DATA":
				inData = true
				writeLine("354 End data")
			case upper == "QUIT":
				writeLine("221 Bye")
				return
			default:
				writeLine("250 Ok")
			}
		}
	}()

	return ln.Addr().String(), func() string {
			mu.Lock()
			defer mu.Unlock()
			return dataBuf.String()
		}, func() {
			_ = ln.Close()
			wg.Wait()
		}
}

func startFakeSMTPServer(t *testing.T, tlsCfg *tls.Config) (string, func() string, func()) {
	t.Helper()
	var ln net.Listener
	var err error
	if tlsCfg != nil {
		ln, err = tls.Listen("tcp", "127.0.0.1:0", tlsCfg)
	} else {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
	}
	if err != nil {
		t.Fatalf("failed to listen on fake SMTP port: %v", err)
	}

	var mu sync.Mutex
	var dataBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			return
		}
		defer conn.Close()
		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		writeLine := func(line string) {
			_, _ = rw.WriteString(line + "\r\n")
			_ = rw.Flush()
		}
		writeLine("220 localhost ESMTP FakeSMTP")
		inData := false
		for {
			line, readErr := rw.ReadString('\n')
			if readErr != nil {
				return
			}
			trimmed := strings.TrimRight(line, "\r\n")
			if inData {
				if trimmed == "." {
					inData = false
					writeLine("250 2.0.0 Ok")
					continue
				}
				mu.Lock()
				dataBuf.WriteString(line)
				mu.Unlock()
				continue
			}
			upper := strings.ToUpper(trimmed)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				writeLine("250-localhost")
				writeLine("250 SIZE 10485760")
			case strings.HasPrefix(upper, "MAIL FROM:"), strings.HasPrefix(upper, "RCPT TO:"):
				writeLine("250 2.1.0 Ok")
			case upper == "DATA":
				inData = true
				writeLine("354 End data with <CR><LF>.<CR><LF>")
			case upper == "QUIT":
				writeLine("221 Bye")
				return
			default:
				writeLine("250 Ok")
			}
		}
	}()

	return ln.Addr().String(), func() string {
			mu.Lock()
			defer mu.Unlock()
			return dataBuf.String()
		}, func() {
			_ = ln.Close()
			wg.Wait()
		}
}

func generateSelfSignedLoopbackCert(t *testing.T) tls.Certificate {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa keygen failed: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("x509 CreateCertificate failed: %v", err)
	}
	return tls.Certificate{
		Certificate: [][]byte{der},
		PrivateKey:  priv,
	}
}
