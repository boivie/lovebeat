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
			var ret []model.Service
			if view, ok := servicesState.views[c.View]; ok {
				for _, s := range view.servicesInView {
					ret = append(ret, s.data)
				}
			}
			c.Reply <- ret
		case c := <-svcs.getServiceChan:
			var ret = servicesState.services[c.Name]
			if ret == nil {
				c.Reply <- nil
			} else {
				c.Reply <- &ret.data
			}
		case c := <-svcs.getViewsChan:
			var ret []model.View
			for _, v := range servicesState.views {
				ret = append(ret, v.data)
			}
			c.Reply <- ret
		case c := <-svcs.getViewChan:
			if ret, ok := servicesState.views[c.Name]; ok {
				c.Reply <- &ret.data
			} else {
				c.Reply <- nil
			}
		case c := <-svcs.updateChan:
			if log.IsEnabledFor(logging.DEBUG) {
				j, _ := json.Marshal(c)
				log.Debug("UPDATE: %s", string(j))
			}
			updates := updateServices(servicesState, c)
			updates = updateViews(servicesState, updates)
			printUpdates(updates)
			persist(be, updates)
			sendBusEvents(bus, updates)
			triggerAlerters(alerter, updates)
		}
	}
}
