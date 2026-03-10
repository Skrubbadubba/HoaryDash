// ws_proxy.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

var clientUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsProxyHandler(haBaseURL, haToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientConn, err := clientUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("ws upgrade error:", err)
			return
		}
		log.Printf("Client connected from %s", clientConn.RemoteAddr())
		defer clientConn.Close()

		if haBaseURL == "" {
			log.Print("HA url not set, defaulting to 'http://homeassistant.local:8123'")
			haBaseURL = "http://homeassistant.local:8123"
		}

		haURL, _ := url.Parse(haBaseURL)
		haURL.Scheme = "ws"
		haURL.Path = "/api/websocket"

		haConn, _, err := websocket.DefaultDialer.Dial(haURL.String(), nil)
		if err != nil {
			log.Println("ws dial HA error:", err)
			log.Printf("Tried dialing %v", haURL)
			return
		}
		log.Printf("Connected to ha ws at %s", haURL.String())
		defer haConn.Close()

		if haToken == "" {
			log.Print("Getting HA token fron environment")
			envToken := os.Getenv("HA_TOKEN")
			if envToken == "" {
				log.Printf("No HA token could be read")
				return
			}
			haToken = envToken
		}
		if err := haAuth(haConn, haToken); err != nil {
			log.Println("ws HA auth error:", err)
			return
		}

		errc := make(chan error, 2)

		go func() {
			// from HA → send to client
			for {
				mt, msg, err := haConn.ReadMessage()
				if err != nil {
					errc <- err
					return
				}
				log.Printf("Recieved ws message from HA: \n%s", msg)
				if err := clientConn.WriteMessage(mt, msg); err != nil {
					errc <- err
					return
				}
			}
		}()

		go func() {
			// from client → send to HA
			for {
				mt, msg, err := clientConn.ReadMessage()
				if err != nil {
					errc <- err
					return
				}
				log.Printf("Recieved ws message from client: \n%s", msg)
				if err := haConn.WriteMessage(mt, msg); err != nil {
					errc <- err
					return
				}
			}
		}()

		<-errc // block until one side closes
	}
}

// haAuth performs the HA WebSocket auth handshake.
// HA sends auth_required → we send auth → HA sends auth_ok.
func haAuth(conn *websocket.Conn, token string) error {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg, &envelope); err != nil {
		return err
	}

	auth := map[string]string{"type": "auth", "access_token": token}
	if err := conn.WriteJSON(auth); err != nil {
		return err
	}
	jsonString, _ := json.Marshal(auth)
	log.Printf("Sent auth json: %s", jsonString)

	_, msg, err = conn.ReadMessage()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(msg, &envelope); err != nil {
		return err
	}
	if envelope.Type != "auth_ok" {
		return fmt.Errorf("HA auth failed: %s", envelope.Type)
	}
	log.Printf("Got message: %s", msg)

	return nil
}
