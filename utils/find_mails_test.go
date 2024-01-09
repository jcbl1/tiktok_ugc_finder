package utils

import (
	"net/mail"
	"testing"
)

var body string = `abc@slkdfh.com
sdlkfh@sdkf.sdhkf.skldfh.de 😭`

func TestFindMails(t *testing.T) {
	var mails []*mail.Address
	if err := FindMails(body, &mails); err != nil {
		t.Error(err)
	} else {
		t.Log(mails)
	}
	mails = mails[:0]
	if err := FindMails("", &mails); err != nil {
		t.Error(err)
	} else {
		t.Log(mails)
	}
}
