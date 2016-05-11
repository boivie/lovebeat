package model

// ViewAddedEvent will be sent when a view has been added
type ViewAddedEvent struct {
	View View `json:"view"`
}

// ViewStateChangedEvent will be sent when a view has changed state
type ViewStateChangedEvent struct {
	View           View      `json:"view"`
	Previous       string    `json:"previous"`
	Current        string    `json:"current"`
	FailedServices []Service `json:"failed"`
}

// ServiceStateChangedEvent will be sent when a service has changed state
type ServiceStateChangedEvent struct {
	Service  Service `json:"service"`
	Previous string  `json:"previous"`
	Current  string  `json:"current"`
}

// When adding a new expression struct type here, don't forget
// to add it to the test cases so the member names are checked
// for conformity.

// ServiceAddedEvent will be sent when a service has been added
type ServiceAddedEvent struct {
	Service Service `json:"service"`
}

// ServiceRemovedEvent will be sent when a service has been removed
type ServiceRemovedEvent struct {
	Service Service `json:"service"`
}
