package worker

import (
	"github.com/matheusparro/shorty/internal/queue"
)

// cada consumer registra uma função aqui
var registrations []func(c *queue.Consumer)

// register é usado pelos arquivos clicks.go/created.go/... via init()
func register(fn func(c *queue.Consumer)) {
	registrations = append(registrations, fn)
}

// applyRegistrations aplica tudo no consumer (chamado pelo worker.Run)
func applyRegistrations(c *queue.Consumer) {
	for _, fn := range registrations {
		fn(c)
	}
}
