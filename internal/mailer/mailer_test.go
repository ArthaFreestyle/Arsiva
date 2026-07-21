package mailer

import (
	"net/mail"
	"strings"
	"testing"
)

func testMailer() *smtpMailer {
	return &smtpMailer{cfg: Config{From: "admin@arsiva.id", FromName: "Arsiva"}}
}

// Gmail rejects messages without a Message-ID at end-of-DATA ("550-5.7.1
// Messages missing a valid Message-ID header are not accepted"), and Date is
// equally mandatory per RFC 5322. Both builders must emit them.
func TestMessagesCarryRequiredHeaders(t *testing.T) {
	m := testMailer()

	cases := map[string][]byte{
		"plain":     m.buildMessage("student@gmail.com", "Kode Verifikasi", "123456"),
		"multipart": m.buildMultipartMessage("student@gmail.com", "Kode Verifikasi", "<p>123456</p>", "123456"),
	}

	for name, raw := range cases {
		msg, err := mail.ReadMessage(strings.NewReader(string(raw)))
		if err != nil {
			t.Fatalf("%s: message is not parseable as RFC 5322: %v", name, err)
		}

		id := msg.Header.Get("Message-ID")
		if !strings.HasPrefix(id, "<") || !strings.HasSuffix(id, ">") {
			t.Errorf("%s: Message-ID must be angle-bracketed, got %q", name, id)
		}
		if !strings.HasSuffix(id, "@arsiva.id>") {
			t.Errorf("%s: Message-ID should use the sender domain, got %q", name, id)
		}

		if _, err := msg.Header.Date(); err != nil {
			t.Errorf("%s: Date header missing or unparseable: %v", name, err)
		}
	}
}

func TestMessageIDIsUnique(t *testing.T) {
	m := testMailer()
	if a, b := m.newMessageID(), m.newMessageID(); a == b {
		t.Errorf("consecutive Message-IDs collided: %q", a)
	}
}
