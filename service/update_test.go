package service

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/model"
	"regexp"
	"testing"
)

func TestNewService(t *testing.T) {
	state := newState()
	updates := updateServices(state, &model.Update{Ts: 1000, Service: "test", Beat: &model.Beat{}})

	if state.services["test"].lastBeat != 1000 {
		t.Errorf("Missing beat")
	}

	if updates[0].oldService != nil || updates[0].newService.data.Name != "test" {
		t.Errorf("Missing update")
	}
}

func TestExpireService(t *testing.T) {
	state := newState()
	updateServices(state, &model.Update{Ts: 0, Service: "test", SetTimeout: &model.SetTimeout{Timeout: 1000}, Beat: &model.Beat{}})
	updates := updateServices(state, &model.Update{Ts: 1000, Tick: &model.Tick{}})

	if updates[0].oldService.name() != "test" || updates[0].oldService.data.State != model.StateOk || updates[0].newService.data.State != model.StateError {
		t.Errorf("Expected service in error")
	}

	updates2 := updateServices(state, &model.Update{Ts: 2000, Tick: &model.Tick{}})

	if len(updates2) != 0 {
		t.Errorf("Expected no more errors")
	}
}

func TestFastBeatsRateLimited(t *testing.T) {
	state := newState()
	updateServices(state, &model.Update{Ts: 0, Service: "test", SetTimeout: &model.SetTimeout{Timeout: 10000}, Beat: &model.Beat{}})
	updates := updateServices(state, &model.Update{Ts: 1000, Service: "test", Beat: &model.Beat{}})
	if len(updates) == 0 {
		t.Errorf("Expected updates")
	}

	updates2 := updateServices(state, &model.Update{Ts: 1100, Service: "test", Beat: &model.Beat{}})

	if len(updates2) != 0 {
		t.Errorf("Expected no updates within the same second")
	}

	updates3 := updateServices(state, &model.Update{Ts: 2000, Service: "test", Beat: &model.Beat{}})

	if len(updates3) == 0 {
		t.Errorf("Expected updates")
	}

	updates4 := updateServices(state, &model.Update{Ts: 2900, Service: "test", Beat: &model.Beat{}})

	if len(updates4) != 0 {
		t.Errorf("Expected no updates within the same second")
	}
}

func TestMuteExpired(t *testing.T) {
	state := newState()
	updateServices(state, &model.Update{Ts: 0, Service: "test", SetTimeout: &model.SetTimeout{Timeout: 10000}, Beat: &model.Beat{}})
	updates1 := updateServices(state, &model.Update{Ts: 5000, Service: "test", MuteService: &model.MuteService{Muted: true}})

	if updates1[0].oldService.name() != "test" || updates1[0].oldService.data.State != model.StateOk || updates1[0].newService.data.State != model.StateMuted {
		t.Errorf("Expected service muted")
	}

	// Idempotent muting
	updates2 := updateServices(state, &model.Update{Ts: 5000, Service: "test", MuteService: &model.MuteService{Muted: true}})

	if len(updates2) != 0 {
		t.Errorf("Expected no updates")
	}

	updates3 := updateServices(state, &model.Update{Ts: 5000, Tick: &model.Tick{}})

	if len(updates3) != 0 {
		t.Errorf("Expected no errors")
	}

	// un-mute
	updates4 := updateServices(state, &model.Update{Ts: 12000, Service: "test", MuteService: &model.MuteService{Muted: false}})
	if updates4[0].oldService.name() != "test" || updates4[0].oldService.data.State != model.StateMuted || updates4[0].newService.data.State != model.StateError {
		t.Errorf("Expected service in error")
	}

	// Idempotent un-muting
	updates5 := updateServices(state, &model.Update{Ts: 12000, Service: "test", MuteService: &model.MuteService{Muted: false}})
	if len(updates5) != 0 {
		t.Errorf("Expected no updates")
	}
}

func TestDeleteService(t *testing.T) {
	state := newState()
	updateServices(state, &model.Update{Ts: 0, Service: "test", SetTimeout: &model.SetTimeout{Timeout: 1000}, Beat: &model.Beat{}})

	s, exists := state.services["test"]
	if s == nil || !exists {
		t.Errorf("Expected service to be present in state")
	}

	updates1 := updateServices(state, &model.Update{Ts: 1000, Service: "test", DeleteService: &model.DeleteService{}})

	if updates1[0].oldService.name() != "test" || updates1[0].oldService.data.State != model.StateOk || updates1[0].newService != nil {
		t.Errorf("Expected service to be deleted")
	}

	s, exists = state.services["test"]
	if s != nil || exists {
		t.Errorf("Expected service to not be present in state")
	}

	// Idempotent deletions
	updates2 := updateServices(state, &model.Update{Ts: 1000, Service: "test", DeleteService: &model.DeleteService{}})

	if len(updates2) != 0 {
		t.Errorf("Expected no updates")
	}
}

func TestSimpleFromTemplate(t *testing.T) {
	state := newState()
	state.alarmStates = []*model.Alarm{&model.Alarm{Name: "testalarm", IncidentNbr: 4}}
	state.alarmTemplates = []alarmTemplate{alarmTemplate{
		config:   config.ConfigAlarm{Name: "testalarm"},
		includes: []*regexp.Regexp{regexp.MustCompile(makePattern("test.*"))},
	}}

	updates1 := updateServices(state, &model.Update{Ts: 0, Service: "test.service", SetTimeout: &model.SetTimeout{Timeout: 1000}, Beat: &model.Beat{}})
	updates1 = updateAlarms(state, 0, updates1)

	if state.alarms["testalarm"].data.State != model.StateOk {
		t.Errorf("Expected alarm OK")
	}

	if updates1[0].oldAlarm != nil || updates1[0].newAlarm.data.Name != "testalarm" {
		t.Errorf("Expected alarm update")
	}

	updates2 := updateServices(state, &model.Update{Ts: 1000, Tick: &model.Tick{}})
	updates2 = updateAlarms(state, 0, updates2)

	if updates2[0].oldService.name() != "test.service" {
		t.Errorf("Expected service in update")
	}

	if updates2[1].oldAlarm.name() != "testalarm" || updates2[1].oldAlarm.data.State != model.StateOk || updates2[1].newAlarm.data.State != model.StateError {
		t.Errorf("Expected alarm in update")
	}

	if updates2[1].newAlarm.data.IncidentNbr != 5 {
		t.Errorf("Expected increase of incident number")
	}

	if state.alarms["testalarm"].data.State != model.StateError {
		t.Errorf("Expected alarm in error")
	}

}

func TestDeleteServiceInAlarm(t *testing.T) {
	state := newState()
	state.alarmStates = []*model.Alarm{&model.Alarm{Name: "testalarm", IncidentNbr: 4}}
	state.alarmTemplates = []alarmTemplate{alarmTemplate{
		config:   config.ConfigAlarm{Name: "testalarm"},
		includes: []*regexp.Regexp{regexp.MustCompile(makePattern("test.*"))},
	}}

	updates1 := updateServices(state, &model.Update{Ts: 0, Service: "test.service", SetTimeout: &model.SetTimeout{Timeout: 1000}, Beat: &model.Beat{}})
	updates1 = updateAlarms(state, 0, updates1)

	if state.alarms["testalarm"].data.State != model.StateOk {
		t.Errorf("Expected alarm OK")
	}

	if updates1[0].oldAlarm != nil || updates1[0].newAlarm.data.Name != "testalarm" {
		t.Errorf("Expected alarm update")
	}

	if state.alarms["testalarm"].servicesInAlarm[0] != state.services["test.service"] {
		t.Errorf("Expected service in alarm")
	}

	updates2 := updateServices(state, &model.Update{Ts: 1000, Tick: &model.Tick{}})
	updates2 = updateAlarms(state, 0, updates2)

	if updates2[0].oldService.name() != "test.service" {
		t.Errorf("Expected service in update")
	}

	if updates2[1].oldAlarm.name() != "testalarm" || updates2[1].oldAlarm.data.State != model.StateOk || updates2[1].newAlarm.data.State != model.StateError {
		t.Errorf("Expected alarm in update")
	}

	if state.alarms["testalarm"].data.State != model.StateError {
		t.Errorf("Expected alarm in error")
	}

	updates3 := updateServices(state, &model.Update{Ts: 0, Service: "test.service", DeleteService: &model.DeleteService{}})
	updates3 = updateAlarms(state, 0, updates3)

	if updates3[0].oldService.name() != "test.service" || updates3[0].newService != nil {
		t.Errorf("Expected service in update")
	}

	if updates3[1].oldAlarm.name() != "testalarm" || updates3[1].oldAlarm.data.State != model.StateError || updates3[1].newAlarm.data.State != model.StateOk {
		t.Errorf("Expected alarm in update")
	}

	if state.alarms["testalarm"].data.State != model.StateOk {
		t.Errorf("Expected alarm in error")
	}

	if len(state.alarms["testalarm"].servicesInAlarm) != 0 {
		t.Errorf("Expected service removed from alarm")
	}
}

func TestDeleteAlarm(t *testing.T) {
	state := newState()
	state.alarmStates = []*model.Alarm{&model.Alarm{Name: "testalarm", IncidentNbr: 4}}
	state.alarmTemplates = []alarmTemplate{alarmTemplate{
		config:   config.ConfigAlarm{Name: "testalarm"},
		includes: []*regexp.Regexp{regexp.MustCompile(makePattern("test.*"))},
	}}

	updates1 := updateServices(state, &model.Update{Ts: 0, Service: "test.service", SetTimeout: &model.SetTimeout{Timeout: 1000}, Beat: &model.Beat{}})
	updates1 = updateAlarms(state, 0, updates1)

	u := &model.Update{Ts: 1, Alarm: "testalarm", DeleteAlarm: &model.DeleteAlarm{}}
	updates2 := updateServices(state, u)
	updates2 = removeAlarms(state, u, updates2)
	updates2 = updateAlarms(state, 0, updates2)

	if state.alarms["testalarm"].data.State != model.StateOk {
		t.Errorf("Expected alarm still in state")
	}

	updates3 := updateServices(state, &model.Update{Ts: 0, Service: "test.service", DeleteService: &model.DeleteService{}})
	updateAlarms(state, 0, updates3)

	u2 := &model.Update{Ts: 1, Alarm: "testalarm", DeleteAlarm: &model.DeleteAlarm{}}
	updates4 := updateServices(state, u2)
	updates4 = removeAlarms(state, u2, updates4)
	updates4 = updateAlarms(state, 0, updates4)

	_, exists := state.alarms["testalarm"]
	if exists {
		t.Errorf("Expected alarm removed from state")
	}
}

func TestInitialInError(t *testing.T) {
	state := newState()
	state.alarmStates = make([]*model.Alarm, 0)
	state.alarmTemplates = []alarmTemplate{alarmTemplate{
		config:   config.ConfigAlarm{Name: "testalarm"},
		includes: []*regexp.Regexp{regexp.MustCompile(makePattern("test.*"))},
	}}

	updates1 := updateServices(state, &model.Update{Ts: 0, Service: "test.service", SetTimeout: &model.SetTimeout{Timeout: 0}, Beat: &model.Beat{}})
	updates1 = updateAlarms(state, 0, updates1)

	if state.alarms["testalarm"].data.State != model.StateError {
		t.Errorf("Expected alarm in error (was: %v)", state.alarms["testalarm"].data.State)
	}
	if state.alarms["testalarm"].data.IncidentNbr != 1 {
		t.Errorf("Expected incident #1 (was: %v)", state.alarms["testalarm"].data.IncidentNbr)
	}

	if updates1[0].oldAlarm != nil || updates1[0].newAlarm.data.Name != "testalarm" || updates1[0].newAlarm.data.State != model.StateNew {
		t.Errorf("Expected new alarm as New")
	}
	if updates1[2].oldAlarm.data.Name != "testalarm" || updates1[2].oldAlarm.data.State != model.StateNew || updates1[2].newAlarm.data.State != model.StateError {
		t.Errorf("Expected updated alarm in error")
	}
}
