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

var (
	EMPTY_REGEXP = regexp.MustCompile("^$")
)

func (v *View) name() string { return v.data.Name }
func (v *View) update(ts int64) {
	v.data.State = model.STATE_OK
	v.data.LastUpdated = ts
	for _, s := range v.services {
		if v.contains(s.name()) {
			if s.data.State == model.STATE_WARNING &&
				v.data.State == model.STATE_OK {
				v.data.State = model.STATE_WARNING
			} else if s.data.State == model.STATE_ERROR {
				v.data.State = model.STATE_ERROR
			}
		}
	}
}

func (v *View) contains(serviceName string) bool {
	return v.ree.Match([]byte(serviceName))
}

func (v *View) save(be backend.Backend, ref *View, ts int64) {
	if v.data.State != ref.data.State && ref.data.State == model.STATE_OK {
		v.data.IncidentNbr += 1
	}
	be.SaveView(&v.data)
}
