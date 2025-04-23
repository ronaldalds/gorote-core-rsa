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

func (con *Controller) websocketHandler(ctx *websocket.Conn) {
	// Obter dados do contexto
	req := ctx.Locals("validatedData").(*WsConn)
	clientID := req.ID

	// Registrar conexão
	wsMux.Lock()
	Cws[clientID] = ctx
	wsMux.Unlock()
	log.Printf("Novo cliente conectado: %v", clientID)

	defer func() {
		// Remover conexão ao desconectar
		wsMux.Lock()
		delete(Cws, clientID)
		wsMux.Unlock()
		log.Printf("Cliente desconectado: %v", clientID)
		ctx.Close()
	}()

	// Enviar mensagem de boas-vindas
	welcomeMsg := map[string]any{
		"type":    "welcome",
		"message": "Bem-vindo ao WebSocket seguro!",
		"sub":     clientID,
	}
	if err := ctx.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Erro ao enviar mensagem de boas-vindas: %v", err)
		return
	}

	// Loop para lidar com mensagens
	for {
		var msg map[string]any
		err := ctx.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Erro na leitura: %v", err)
			}
			break
		}

		log.Printf("Mensagem recebida de %v: %v", clientID, msg)
	}
}

// GetConnection obtém uma conexão de forma segura
func GetConnection(id uint) (*websocket.Conn, bool) {
	wsMux.RLock()
	defer wsMux.RUnlock()
	conn, exists := Cws[id]
	return conn, exists
}

// Broadcast envia mensagem para todos os clientes
func Broadcast(message any) {
	wsMux.RLock()
	defer wsMux.RUnlock()

	for id, conn := range Cws {
		go func(id uint, c *websocket.Conn) {
			if err := c.WriteJSON(message); err != nil {
				log.Printf("Erro ao enviar para cliente %d: %v", id, err)
				// Opcional: remover conexão problemática
				wsMux.Lock()
				delete(Cws, id)
				wsMux.Unlock()
			}
		}(id, conn)
	}
}

// SendTo envia mensagem para um cliente específico
func SendTo(id uint, message any) error {
	wsMux.RLock()
	conn, exists := Cws[id]
	wsMux.RUnlock()

	if !exists {
		return errors.New("cliente não encontrado")
	}

	return conn.WriteJSON(message)
}
