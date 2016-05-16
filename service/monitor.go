package service

import (
	"encoding/json"
	"github.com/boivie/lovebeat/alert"
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/eventbus"
	"github.com/boivie/lovebeat/model"
	"github.com/boivie/lovebeat/notify"
	"github.com/op/go-logging"
	"time"
)

func (svcs *ServicesImpl) Monitor(cfg config.Config, notifier notify.Notifier, be backend.Backend, bus *eventbus.EventBus, alerter alert.Alerter) {
	servicesState := newState()
	loadState(servicesState, be, cfg)

	notifyTimer := time.NewTicker(time.Duration(60) * time.Second)

	for {
		select {
		case <-notifyTimer.C:
			notifier.Notify("monitor")
		case c := <-svcs.getServicesChan:
			ret := make([]model.Service, 0)
			if view, ok := servicesState.views[c.View]; ok {
				for _, s := range view.servicesInView {
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
		case c := <-svcs.getViewsChan:
			ret := make([]model.View, 0)
			for _, v := range servicesState.views {
				ret = append(ret, v.getExternalModel())
			}
			c.Reply <- ret
		case c := <-svcs.getViewChan:
			if ret, ok := servicesState.views[c.Name]; ok {
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
			updates = updateViews(servicesState, c.Ts, updates)
			printUpdates(updates)
			persist(be, updates)
			sendBusEvents(bus, updates)
			triggerAlerters(alerter, updates)
		}
	}
}
