package wshandler

import (
	"log"

	"github.com/asek-ll/aecc-server/internal/ws"
)

var _ ws.Handler = &Handler{}

type Handler struct {
}

func (h *Handler) HandleMessage(content []byte, clientId uint) error {
	log.Println(content)
	return nil
}
