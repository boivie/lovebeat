package alert

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/model"
	"strings"
	"testing"
)

const email = "test@example.com"

func TestSimpleSubject(t *testing.T) {
	v1 := model.Alarm{Name: "TestAlarm", State: "ok", IncidentNbr: 1}
	alert := AlertInfo{v1, "ok", "error", config.ConfigAlarm{}}
	mail := createMail(email, alert)
	if mail.Subject != "[LOVEBEAT] TestAlarm-1" {
		t.Errorf("Was: %v", mail.Subject)
	}
}

func TestSimpleBody(t *testing.T) {
	v1 := model.Alarm{Name: "TestAlarm", State: "ok", IncidentNbr: 1}
	alert := AlertInfo{v1, "ok", "error", config.ConfigAlarm{}}
	mail := createMail(email, alert)
	ref := "The status for alarm 'TestAlarm' has changed from 'OK' to 'ERROR'"

	if !strings.Contains(mail.Body, ref) {
		t.Errorf("Was: %v", mail.Body)
	}
}
