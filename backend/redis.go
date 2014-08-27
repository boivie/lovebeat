package backend

import (
	"encoding/json"
	"github.com/hoisie/redis"
	"github.com/op/go-logging"
)

var (
	flog = logging.MustGetLogger("lovebeat")
)

type RedisBackend struct {
	client redis.Client
}

func (r RedisBackend) loadView(name string) *StoredView {
	view := &StoredView{
		Name:   name,
		State:  STATE_OK,
		Regexp: "^$",
	}

	if data, err := r.client.Get("lb.view." + name); err == nil {
		json.Unmarshal(data, &view)
	}
	return view
}

func (r RedisBackend) SaveService(service *StoredService) {
	b, _ := json.Marshal(service)
	r.client.Set("lb.service."+service.Name, b)
	r.client.Sadd("lb.services.all", []byte(service.Name))
}

func (r RedisBackend) SaveView(view *StoredView) {
	b, _ := json.Marshal(view)
	r.client.Set("lb.view."+view.Name, b)
	r.client.Sadd("lb.views.all", []byte(view.Name))
}

func (r RedisBackend) loadService(name string) *StoredService {
	service := &StoredService{
		Name:           name,
		LastValue:      -1,
		LastBeat:       -1,
		LastUpdated:    -1,
		WarningTimeout: -1,
		ErrorTimeout:   -1,
		State:          STATE_PAUSED,
	}

	if data, err := r.client.Get("lb.service." + name); err == nil {
		json.Unmarshal(data, &service)
	}
	return service
}

func (r RedisBackend) LoadServices() []*StoredService {
	var ret []*StoredService

	var namesBytes, _ = r.client.Smembers("lb.services.all")
	for _, nameByte := range namesBytes {
		var name = string(nameByte)
		ret = append(ret, r.loadService(name))
		flog.Debug("Found service %s", name)
	}
	return ret
}

func (r RedisBackend) LoadViews() []*StoredView {
	var ret []*StoredView

	var namesBytes, _ = r.client.Smembers("lb.views.all")
	for _, nameByte := range namesBytes {
		var name = string(nameByte)
		ret = append(ret, r.loadView(name))
		flog.Debug("Found view %s", name)
	}

	return ret
}

func (r RedisBackend) DeleteService(name string) {
	r.client.Srem("lb.services.all", []byte(name))
}

func (r RedisBackend) DeleteView(name string) {
	r.client.Srem("lb.views.all", []byte(name))
}
