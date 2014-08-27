package backend

import (
	"encoding/json"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	log = logging.MustGetLogger("lovebeat")
)

type FileBackend struct {
	path string
}

func (r FileBackend) viewFile(name string) string {
	return filepath.Join(r.path, "views", name+".json")
}
func (r FileBackend) serviceFile(name string) string {
	return filepath.Join(r.path, "services", name+".json")
}

func (r FileBackend) readViewFile(name string, fname string) *StoredView {
	view := &StoredView{
		Name:   name,
		State:  STATE_OK,
		Regexp: "^$",
	}

	data, err := ioutil.ReadFile(fname)
	if err == nil {
		json.Unmarshal(data, &view)
	}

	return view
}

func (r FileBackend) loadView(name string) *StoredView {
	return r.readViewFile(name, r.viewFile(name))
}

func (r FileBackend) SaveService(service *StoredService) {
	b, _ := json.Marshal(service)
	ioutil.WriteFile(r.serviceFile(service.Name), b, 0644)
}

func (r FileBackend) SaveView(view *StoredView) {
	b, _ := json.Marshal(view)
	ioutil.WriteFile(r.viewFile(view.Name), b, 0644)
}

func (r FileBackend) readServiceFile(name string, fname string) *StoredService {
	service := &StoredService{
		Name:           name,
		LastValue:      -1,
		LastBeat:       -1,
		LastUpdated:    -1,
		WarningTimeout: -1,
		ErrorTimeout:   -1,
		State:          STATE_PAUSED,
	}

	data, err := ioutil.ReadFile(fname)
	if err == nil {
		json.Unmarshal(data, &service)
	}

	return service
}

func (r FileBackend) loadService(name string) *StoredService {
	return r.readServiceFile(name, r.serviceFile(name))
}

func (r FileBackend) LoadServices() []*StoredService {
	var ret []*StoredService

	matches, _ := filepath.Glob(filepath.Join(r.path, "services", "*.json"))
	for _, fname := range matches {
		svc := r.readServiceFile("", fname)
		ret = append(ret, svc)
		log.Debug("Found service %s", svc.Name)
	}
	return ret
}

func (r FileBackend) LoadViews() []*StoredView {
	var ret []*StoredView

	matches, _ := filepath.Glob(filepath.Join(r.path, "views", "*.json"))
	for _, fname := range matches {
		view := r.readViewFile("", fname)
		ret = append(ret, view)
		log.Debug("Found view %s", view.Name)
	}
	return ret
}

func (r FileBackend) DeleteService(name string) {
	os.Remove(r.serviceFile(name))
}

func (r FileBackend) DeleteView(name string) {
	os.Remove(r.viewFile(name))
}

func NewFileBackend(path string) Backend {
	os.MkdirAll(filepath.Join(path, "views"), 0755)
	os.MkdirAll(filepath.Join(path, "services"), 0755)
	return &FileBackend{path: path}
}
