// server/routes.go
package server

func (s *Server) routes() {
	// Our first route: the home page that lists all devices.
	s.router.HandleFunc("/", s.handleDevicesList()).Methods("GET")

	s.router.HandleFunc("/devices/add", s.handleDeviceAddForm()).Methods("GET")
	s.router.HandleFunc("/devices/add", s.handleDeviceAddSubmit()).Methods("POST")

	s.router.HandleFunc("/devices/edit/{host}", s.handleDeviceEditForm()).Methods("GET")
	s.router.HandleFunc("/devices/edit/{host}", s.handleDeviceEditSubmit()).Methods("POST")
	s.router.HandleFunc("/devices/remove/{host}", s.handleDeviceRemove()).Methods("POST")
}
