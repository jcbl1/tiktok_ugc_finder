// Package utils implements some utilities.
package utils

import (
	"net/mail"
	"regexp"
	"strings"
)

// FindMails finds mails contained in str. If there are patterns LIKE an email address, they are appended to mails.
func FindMails(str string, mails *[]*mail.Address) error {
	re := regexp.MustCompile(`[a-zA-Z0-9\-]+[a-z0-9._%+\-]*@[a-zA-Z0-9\-]+[a-z0-9.\-]*\.[a-z]{2,4}`)
	matches := re.FindAllString(str, -1)
	for _, match := range matches {
		match = strings.ToLower(match)
		// log.Println("✉️: match:",match)
		m, err := mail.ParseAddress(match)
		if err != nil {
			return err
		}
		*mails = append(*mails, m)
	}

	return nil
}
