package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var clientUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type wsMsg struct {
	mt   int
	data []byte
}

var warmingPool = struct {
	sync.Mutex
	conns map[string][]*warmingConn
}{conns: map[string][]*warmingConn{}}

type warmingConn struct {
	haConn       *websocket.Conn
	dialed       chan struct{}
	stop         chan struct{}
	initialState []byte
	stateDone    chan struct{}
}

func warm(ip, haWsURL, haToken string, subscribeMsg []byte) {
	wc := &warmingConn{
		dialed:    make(chan struct{}),
		stop:      make(chan struct{}),
		stateDone: make(chan struct{}),
	}

	warmingPool.Lock()
	warmingPool.conns[ip] = append(warmingPool.conns[ip], wc)
	warmingPool.Unlock()

	go func() {
		haConn, _, err := websocket.DefaultDialer.Dial(haWsURL, nil)
		if err != nil {
			close(wc.dialed)
			close(wc.stateDone)
			return
		}
		wc.haConn = haConn
		close(wc.dialed)

		if err := haAuth(haConn, haToken); err != nil {
			haConn.Close()
			close(wc.stateDone)
			return
		}
		if err := haConn.WriteMessage(websocket.TextMessage, subscribeMsg); err != nil {
			haConn.Close()
			close(wc.stateDone)
			return
		}

		for {
			select {
			case <-wc.stop:
				close(wc.stateDone)
				return
			default:
			}
			_, data, err := haConn.ReadMessage()
			if err != nil {
				close(wc.stateDone)
				return
			}
			var env struct {
				Type string `json:"type"`
			}
			json.Unmarshal(data, &env)
			if env.Type == "event" {
				wc.initialState = data
				close(wc.stateDone)
				return
			}
		}
	}()
}

func consumeWarming(ip string) *warmingConn {
	warmingPool.Lock()
	defer warmingPool.Unlock()
	log.Printf("consumeWarming: pool state: %v keys, conns for %s: %d", len(warmingPool.conns), ip, len(warmingPool.conns[ip]))
	conns := warmingPool.conns[ip]
	if len(conns) == 0 {
		return nil
	}
	wc := conns[0]
	warmingPool.conns[ip] = conns[1:]
	return wc
}

func dashFileHandler(inner http.Handler, haWsURL, haToken string, subscribeMsg []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := ipOnly(r.RemoteAddr)
		log.Printf("dashFileHandler: request from %s, ip=%s", r.RemoteAddr, ip)
		warm(ip, haWsURL, haToken, subscribeMsg)
		inner.ServeHTTP(w, r)
	}
}

func wsProxyHandler(haWsURL, haToken string, subscribeMsg []byte, rebuildChan <-chan struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientConn, err := clientUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer clientConn.Close()

		var haConn *websocket.Conn
		var initialState []byte

		ip := ipOnly(clientConn.RemoteAddr().String())
		wc := consumeWarming(ip)

		if wc != nil {
			select {
			case <-wc.dialed:
			}

			if wc.haConn != nil {
				// Tell warm goroutine to stop reading, take over
				close(wc.stop)
				// Grab initial state if warm goroutine got it
				<-wc.stateDone
				haConn = wc.haConn
				initialState = wc.initialState
			}
		}

		if haConn == nil {
			// Cold connect
			haConn, _, err = websocket.DefaultDialer.Dial(haWsURL, nil)
			if err != nil {
				return
			}
			if err := haAuth(haConn, haToken); err != nil {
				haConn.Close()
				return
			}
			if err := haConn.WriteMessage(websocket.TextMessage, subscribeMsg); err != nil {
				haConn.Close()
				return
			}
		}
		defer haConn.Close()

		errc := make(chan error, 2)
		send := make(chan wsMsg, 8)
		done := make(chan struct{})

		go func() {
			for {
				select {
				case msg := <-send:
					if err := clientConn.WriteMessage(msg.mt, msg.data); err != nil {
						errc <- err
						return
					}
				case <-done:
					return
				}
			}
		}()

		if initialState != nil {
			send <- wsMsg{websocket.TextMessage, initialState}
		}

		// HA → client
		go func() {
			for {
				mt, msg, err := haConn.ReadMessage()
				if err != nil {
					errc <- err
					return
				}
				select {
				case send <- wsMsg{mt, msg}:
				case <-done:
					return
				}
			}
		}()

		// client → HA
		go func() {
			for {
				mt, msg, err := clientConn.ReadMessage()
				if err != nil {
					errc <- err
					return
				}
				if err := haConn.WriteMessage(mt, msg); err != nil {
					errc <- err
					return
				}
			}
		}()

		// rebuild → client
		go func() {
			for range rebuildChan {
				select {
				case send <- wsMsg{websocket.TextMessage, []byte("reload")}:
				case <-done:
					return
				}
			}
		}()

		<-errc
		close(done)
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
	if err := conn.WriteJSON(map[string]string{
		"type":         "auth",
		"access_token": token,
	}); err != nil {
		return err
	}
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
	return nil
}

func buildSubscribeMsg(entities map[string]*Entity) []byte {
	ids := make([]string, 0, len(entities))
	for id := range entities {
		ids = append(ids, id)
	}
	msg, _ := json.Marshal(map[string]any{
		"id":         1,
		"type":       "subscribe_entities",
		"entity_ids": ids,
	})
	return msg
}

func handlers(entities map[string]*Entity, haWsURL, haToken string, rebuildChan <-chan struct{}) (fileMiddleware func(http.Handler) http.HandlerFunc, wsHandler http.HandlerFunc) {
	subscribeMsg := buildSubscribeMsg(entities)
	return func(inner http.Handler) http.HandlerFunc {
			return dashFileHandler(inner, haWsURL, haToken, subscribeMsg)
		},
		wsProxyHandler(haWsURL, haToken, subscribeMsg, rebuildChan)
}
