package alert

import (
	"github.com/boivie/lovebeat/model"
	"strings"
	"testing"
)

func TestSimpleSubject(t *testing.T) {
	v1 := model.View{Name: "TestView", State: "ok", IncidentNbr: 1}
	v2 := model.View{Name: "TestView", State: "error", IncidentNbr: 1}
	alert := Alert{v1, v2, []model.Service{}}
	mail := createMail(alert)
	if mail.Subject != "[LOVEBEAT] TestView-1" {
		t.Errorf("Was: %v", mail.Subject)
	}
}

func TestSimpleBody(t *testing.T) {
	v1 := model.View{Name: "TestView", State: "ok", IncidentNbr: 1}
	v2 := model.View{Name: "TestView", State: "error", IncidentNbr: 1}
	alert := Alert{v1, v2, []model.Service{}}
	mail := createMail(alert)
	ref := "The status for view 'TestView' has changed from 'OK' to 'ERROR'"

	if !strings.Contains(mail.Body, ref) {
		t.Errorf("Was: %v", mail.Body)
	}
}

func TestBodyWithServices(t *testing.T) {
	v1 := model.View{Name: "TestView", State: "ok", IncidentNbr: 1}
	v2 := model.View{Name: "TestView", State: "error", IncidentNbr: 1}
	svc1 := model.Service{Name: "Svc1", State: "error"}
	svc2 := model.Service{Name: "Svc2", State: "warning"}
	alert := Alert{v1, v2, []model.Service{svc1, svc2}}
	mail := createMail(alert)

	if !strings.Contains(mail.Body, "* Svc1 - ERROR\n") {
		t.Errorf("Was: %v", mail.Body)
	}

	if !strings.Contains(mail.Body, "* Svc2 - WARNING\n") {
		t.Errorf("Was: %v", mail.Body)
	}

}
