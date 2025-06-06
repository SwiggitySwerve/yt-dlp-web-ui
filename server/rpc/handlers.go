package rpc

import (
	"io"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/rest" // Import for RespondWithErrorJSON
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSockets JSON-RPC handler
func WebSocket(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		rest.RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to upgrade WebSocket connection.", err)
		return
	}

	defer c.Close()

	// notify client that conn is open and ok
	c.WriteJSON(struct{ Status string }{Status: "connected"})

	for {
		mtype, reader, err := c.NextReader()
		if err != nil {
			break
		}

		res := newRequest(reader).Call()

		writer, err := c.NextWriter(mtype)
		if err != nil {
			// Note: It's tricky to send a JSON error response over a WebSocket that's failing at the NextWriter stage.
			// The connection might already be compromised. Logging is important.
			// For consistency, we'll call RespondWithErrorJSON, but it might not reach the client.
			rest.RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to get WebSocket writer.", err)
			break
		}

		io.Copy(writer, res)
	}
}

// HTTP-POST JSON-RPC handler
func Post(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	res := newRequest(r.Body).Call()
	_, err := io.Copy(w, res)

	if err != nil {
		rest.RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to write JSON-RPC response.", err)
		return
	}
}
