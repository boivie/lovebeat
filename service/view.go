package service

import (
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/model"
	"regexp"
)

type View struct {
	services map[string]*Service
	data     model.View
	ree      *regexp.Regexp
}

func (v *View) name() string { return v.data.Name }
func (v *View) update(ts int64) {
	v.data.State = model.StateOk
	for _, s := range v.services {
		if v.contains(s.name()) {
			if s.data.State == model.StateWarning &&
				v.data.State == model.StateOk {
				v.data.State = model.StateWarning
			} else if s.data.State == model.StateError {
				v.data.State = model.StateError
			}
		}
	}
}

func (v *View) contains(serviceName string) bool {
	return v.ree.Match([]byte(serviceName))
}

func (v *View) save(be backend.Backend, ref *View, ts int64) {
	if v.data.State != ref.data.State && ref.data.State == model.StateOk {
		v.data.IncidentNbr += 1
	}
	be.SaveView(&v.data)
}
