package service

import (
	"encoding/json"
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/model"
	"github.com/boivie/lovebeat/notify"
	"github.com/op/go-logging"
	"time"
)

func (svcs *ServicesImpl) Monitor(cfg config.Config, notifier notify.Notifier, be backend.Backend) {
	var observers []ServiceCallback
	servicesState := newState()
	loadState(servicesState, be, cfg)

	notifyTimer := time.NewTicker(time.Duration(60) * time.Second)

	for {
		select {
		case <-notifyTimer.C:
			notifier.Notify("monitor")
		case c := <-svcs.subscribeChan:
			observers = append(observers, c)
		case c := <-svcs.getServicesChan:
			ret := make([]model.Service, 0)
			if c.Alarm != "" {
				if alarm, ok := servicesState.alarms[c.Alarm]; ok {
					for _, s := range alarm.servicesInAlarm {
						ret = append(ret, s.getExternalModel())
					}
				}
			} else {
				for _, s := range servicesState.services {
					ret = append(ret, s.getExternalModel())
				}
			}
			c.Reply <- ret
		case c := <-svcs.getServiceChan:
			var ret = servicesState.services[c.Name]
			if ret == nil {
				c.Reply <- nil
			} else {
				r := ret.getExternalModel()
				c.Reply <- &r
			}
		case c := <-svcs.getAlarmsChan:
			ret := make([]model.Alarm, 0)
			for _, v := range servicesState.alarms {
				ret = append(ret, v.getExternalModel())
			}
			c.Reply <- ret
		case c := <-svcs.getAlarmChan:
			if ret, ok := servicesState.alarms[c.Name]; ok {
				r := ret.getExternalModel()
				c.Reply <- &r
			} else {
				c.Reply <- nil
			}
		case c := <-svcs.updateChan:
			if log.IsEnabledFor(logging.DEBUG) {
				j, _ := json.Marshal(c)
				log.Debugf("UPDATE: %s", string(j))
			}
			updates := updateServices(servicesState, c)
			updates = removeAlarms(servicesState, c, updates)
			updates = updateAlarms(servicesState, c.Ts, updates)
			persist(be, updates)
			sendCallbacks(observers, c, updates)
		}
	}
}
