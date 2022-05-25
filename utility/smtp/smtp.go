package smtp

import (
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"crypto/tls"
	"strings"
)

func SendByTLS(fromEmail string, toEmails []string, subject string, body string, username string, password string, servername string) error {
	from := mail.Address{"", fromEmail}

	headers := map[string]string{}
	headers["From"] = from.String()
	headers["To"] = strings.Join(toEmails, ",")
	headers["Subject"] = subject

	message := ""
	for k,v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", username, password, host)

	tlsconfig := &tls.Config {
		InsecureSkipVerify: true,
		ServerName: host,
	}

	conn, err := tls.Dial("tcp", servername, tlsconfig); if err != nil { return err }

	c, err := smtp.NewClient(conn, host); if err != nil { return err }

	if err = c.Auth(auth); err != nil { return err }

	if err = c.Mail(from.Address); err != nil { return err }

	for _, email := range toEmails {
		err = c.Rcpt(email); if err != nil { return err }
	}

	w, err := c.Data(); if err != nil { return err }

	_, err = w.Write([]byte(message)); if err != nil { return err }

	err = w.Close(); if err != nil { return err }

	c.Quit()
	return nil
}
