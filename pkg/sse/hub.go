package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Event struct {
	ID   string      `json:"id,omitempty"`
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

type Client struct {
	ID       string
	Events   chan Event
	Response http.ResponseWriter
	mu       sync.Mutex
	closed   bool
}

type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				client.Close()
			}
			h.mu.Unlock()

		case event := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Events <- event:
				default:
					go h.Unregister(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(event Event) {
	h.broadcast <- event
}

func NewClient(id string, w http.ResponseWriter) *Client {
	return &Client{
		ID:       id,
		Events:   make(chan Event, 256),
		Response: w,
	}
}

func (c *Client) Send(event Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	fmt.Fprintf(c.Response, "event: %s\n", event.Type)
	
	if event.ID != "" {
		fmt.Fprintf(c.Response, "id: %s\n", event.ID)
	}

	if event.Data != nil {
		data, err := json.Marshal(event.Data)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.Response, "data: %s\n", data)
	}

	fmt.Fprintf(c.Response, "\n")
	
	if flusher, ok := c.Response.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		close(c.Events)
		c.closed = true
	}
}