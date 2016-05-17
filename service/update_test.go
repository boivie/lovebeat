package service

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/model"
	"regexp"
	"testing"
)

func TestNewService(t *testing.T) {
	state := newState()
	updates := updateServices(state, &Update{Ts: 1, Service: "test", Beat: &Beat{}})

	if state.services["test"].lastBeat != 1 {
		t.Errorf("Missing beat")
	}

	if updates[0].oldService != nil || updates[0].newService.data.Name != "test" {
		t.Errorf("Missing update")
	}
}

func TestExpireService(t *testing.T) {
	state := newState()
	updateServices(state, &Update{Ts: 0, Service: "test", SetTimeout: &SetTimeout{Timeout: 1000}, Beat: &Beat{}})
	updates := updateServices(state, &Update{Ts: 1000, Tick: &Tick{}})

	if updates[0].oldService.name() != "test" || updates[0].oldService.data.State != model.StateOk || updates[0].newService.data.State != model.StateError {
		t.Errorf("Expected service in error")
	}

	updates2 := updateServices(state, &Update{Ts: 2000, Tick: &Tick{}})

	if len(updates2) != 0 {
		t.Errorf("Expected no more errors")
	}
}

func TestMuteExpired(t *testing.T) {
	state := newState()
	updateServices(state, &Update{Ts: 0, Service: "test", SetTimeout: &SetTimeout{Timeout: 1000}, Beat: &Beat{}})
	updates1 := updateServices(state, &Update{Ts: 500, Service: "test", MuteService: &MuteService{Muted: true}})

	if updates1[0].oldService.name() != "test" || updates1[0].oldService.data.State != model.StateOk || updates1[0].newService.data.State != model.StateMuted {
		t.Errorf("Expected service muted")
	}

	// Idempotent muting
	updates2 := updateServices(state, &Update{Ts: 500, Service: "test", MuteService: &MuteService{Muted: true}})

	if len(updates2) != 0 {
		t.Errorf("Expected no updates")
	}

	updates3 := updateServices(state, &Update{Ts: 1000, Tick: &Tick{}})

	if len(updates3) != 0 {
		t.Errorf("Expected no errors")
	}

	// un-mute
	updates4 := updateServices(state, &Update{Ts: 1500, Service: "test", MuteService: &MuteService{Muted: false}})
	if updates4[0].oldService.name() != "test" || updates4[0].oldService.data.State != model.StateMuted || updates4[0].newService.data.State != model.StateError {
		t.Errorf("Expected service in error")
	}

	// Idempotent un-muting
	updates5 := updateServices(state, &Update{Ts: 500, Service: "test", MuteService: &MuteService{Muted: false}})
	if len(updates5) != 0 {
		t.Errorf("Expected no updates")
	}
}

func TestDeleteService(t *testing.T) {
	state := newState()
	updateServices(state, &Update{Ts: 0, Service: "test", SetTimeout: &SetTimeout{Timeout: 1000}, Beat: &Beat{}})

	s, exists := state.services["test"]
	if s == nil || !exists {
		t.Errorf("Expected service to be present in state")
	}

	updates1 := updateServices(state, &Update{Ts: 1000, Service: "test", DeleteService: &DeleteService{}})

	if updates1[0].oldService.name() != "test" || updates1[0].oldService.data.State != model.StateOk || updates1[0].newService != nil {
		t.Errorf("Expected service to be deleted")
	}

	s, exists = state.services["test"]
	if s != nil || exists {
		t.Errorf("Expected service to not be present in state")
	}

	// Idempotent deletions
	updates2 := updateServices(state, &Update{Ts: 1000, Service: "test", DeleteService: &DeleteService{}})

	if len(updates2) != 0 {
		t.Errorf("Expected no updates")
	}
}

func TestSimpleFromTemplate(t *testing.T) {
	state := newState()
	state.viewStates = []*model.View{&model.View{Name: "testview", IncidentNbr: 4}}
	state.viewTemplates = []ViewTemplate{ViewTemplate{
		config:   config.ConfigView{Name: "testview"},
		includes: []*regexp.Regexp{regexp.MustCompile(makePattern("test.*"))},
	}}

	updates1 := updateServices(state, &Update{Ts: 0, Service: "test.service", SetTimeout: &SetTimeout{Timeout: 1000}, Beat: &Beat{}})
	updates1 = updateViews(state, 0, updates1)

	if state.views["testview"].data.State != model.StateOk {
		t.Errorf("Expected view OK")
	}

	if updates1[0].oldView != nil || updates1[0].newView.data.Name != "testview" {
		t.Errorf("Expected view update")
	}

	updates2 := updateServices(state, &Update{Ts: 1000, Tick: &Tick{}})
	updates2 = updateViews(state, 0, updates2)

	if updates2[0].oldService.name() != "test.service" {
		t.Errorf("Expected service in update")
	}

	if updates2[1].oldView.name() != "testview" || updates2[1].oldView.data.State != model.StateOk || updates2[1].newView.data.State != model.StateError {
		t.Errorf("Expected view in update")
	}

	if updates2[1].newView.data.IncidentNbr != 5 {
		t.Errorf("Expected increase of incident number")
	}

	if state.views["testview"].data.State != model.StateError {
		t.Errorf("Expected view in error")
	}

}

func TestDeleteServiceInView(t *testing.T) {
	state := newState()
	state.viewStates = []*model.View{&model.View{Name: "testview", IncidentNbr: 4}}
	state.viewTemplates = []ViewTemplate{ViewTemplate{
		config:   config.ConfigView{Name: "testview"},
		includes: []*regexp.Regexp{regexp.MustCompile(makePattern("test.*"))},
	}}

	updates1 := updateServices(state, &Update{Ts: 0, Service: "test.service", SetTimeout: &SetTimeout{Timeout: 1000}, Beat: &Beat{}})
	updates1 = updateViews(state, 0, updates1)

	if state.views["testview"].data.State != model.StateOk {
		t.Errorf("Expected view OK")
	}

	if updates1[0].oldView != nil || updates1[0].newView.data.Name != "testview" {
		t.Errorf("Expected view update")
	}

	if state.views["testview"].servicesInView[0] != state.services["test.service"] {
		t.Errorf("Expected service in view")
	}

	updates2 := updateServices(state, &Update{Ts: 1000, Tick: &Tick{}})
	updates2 = updateViews(state, 0, updates2)

	if updates2[0].oldService.name() != "test.service" {
		t.Errorf("Expected service in update")
	}

	if updates2[1].oldView.name() != "testview" || updates2[1].oldView.data.State != model.StateOk || updates2[1].newView.data.State != model.StateError {
		t.Errorf("Expected view in update")
	}

	if state.views["testview"].data.State != model.StateError {
		t.Errorf("Expected view in error")
	}

	updates3 := updateServices(state, &Update{Ts: 0, Service: "test.service", DeleteService: &DeleteService{}})
	updates3 = updateViews(state, 0, updates3)

	if updates3[0].oldService.name() != "test.service" || updates3[0].newService != nil {
		t.Errorf("Expected service in update")
	}

	if updates3[1].oldView.name() != "testview" || updates3[1].oldView.data.State != model.StateError || updates3[1].newView.data.State != model.StateOk {
		t.Errorf("Expected view in update")
	}

	if state.views["testview"].data.State != model.StateOk {
		t.Errorf("Expected view in error")
	}

	if len(state.views["testview"].servicesInView) != 0 {
		t.Errorf("Expected service removed from view")
	}
}

func TestDeleteView(t *testing.T) {
	state := newState()
	state.viewStates = []*model.View{&model.View{Name: "testview", IncidentNbr: 4}}
	state.viewTemplates = []ViewTemplate{ViewTemplate{
		config:   config.ConfigView{Name: "testview"},
		includes: []*regexp.Regexp{regexp.MustCompile(makePattern("test.*"))},
	}}

	updates1 := updateServices(state, &Update{Ts: 0, Service: "test.service", SetTimeout: &SetTimeout{Timeout: 1000}, Beat: &Beat{}})
	updates1 = updateViews(state, 0, updates1)

	u := &Update{Ts: 1, View: "testview", DeleteView: &DeleteView{}}
	updates2 := updateServices(state, u)
	updates2 = removeViews(state, u, updates2)
	updates2 = updateViews(state, 0, updates2)

	if state.views["testview"].data.State != model.StateOk {
		t.Errorf("Expected view still in state")
	}

	updates3 := updateServices(state, &Update{Ts: 0, Service: "test.service", DeleteService: &DeleteService{}})
	updateViews(state, 0, updates3)

	u2 := &Update{Ts: 1, View: "testview", DeleteView: &DeleteView{}}
	updates4 := updateServices(state, u2)
	updates4 = removeViews(state, u2, updates4)
	updates4 = updateViews(state, 0, updates4)

	_, exists := state.views["testview"]
	if exists {
		t.Errorf("Expected view removed from state")
	}
}
