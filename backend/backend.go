package backend

type Backend interface {
	Start()
}

var be Backend

func Start() {
	if be == nil {
		return
	}
	go be.Start()
}

func SetBackend(b Backend) {
	be = b
}
