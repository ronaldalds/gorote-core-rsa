package example

import (
	"log"

	"github.com/gofiber/contrib/websocket"
)

func (con *appController) websocketHandler(ctx *websocket.Conn) {
	req := ctx.Locals("validatedData").(*WsConn)
	clientID := req.ID

	wsMux.Lock()
	Cws[clientID] = ctx
	wsMux.Unlock()
	log.Printf("Novo cliente conectado: %v", clientID)

	defer func() {
		wsMux.Lock()
		delete(Cws, clientID)
		wsMux.Unlock()
		log.Printf("Cliente desconectado: %v", clientID)
		ctx.Close()
	}()

	welcomeMsg := map[string]any{
		"type":    "welcome",
		"message": "Bem-vindo ao WebSocket seguro!",
		"sub":     clientID,
	}
	if err := ctx.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Erro ao enviar mensagem de boas-vindas: %v", err)
		return
	}

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
