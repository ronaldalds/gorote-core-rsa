package example

import (
	"errors"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type WsConn struct {
	ID uint `param:"id"`
}

var (
	Cws   = make(map[uint]*websocket.Conn)
	wsMux sync.RWMutex
)

func (s *appService) getConnection(id uint) (*websocket.Conn, bool) {
	wsMux.RLock()
	defer wsMux.RUnlock()
	conn, exists := Cws[id]
	return conn, exists
}

func (s *appService) broadcast(message any) {
	wsMux.RLock()
	defer wsMux.RUnlock()

	for id, conn := range Cws {
		go func(id uint, ws *websocket.Conn) {
			if err := ws.WriteJSON(message); err != nil {
				log.Printf("Erro ao enviar para cliente %d: %v", id, err)
				wsMux.Lock()
				delete(Cws, id)
				wsMux.Unlock()
			}
		}(id, conn)
	}
}

func (s *appService) sendTo(id uint, message any) error {
	wsMux.RLock()
	conn, exists := Cws[id]
	wsMux.RUnlock()

	if !exists {
		return errors.New("cliente não encontrado")
	}

	return conn.WriteJSON(message)
}
