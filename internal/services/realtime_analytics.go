package services

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/gorilla/websocket"
)

// RealtimeAnalyticsService handles real-time analytics via WebSocket connections
type RealtimeAnalyticsService struct {
	db              *storage.PostgresStorage
	cache           *storage.RedisStorage
	analyticsService *AnalyticsService
	
	// WebSocket connection management
	clients         map[*websocket.Conn]bool
	clientsMutex    sync.RWMutex
	upgrader        websocket.Upgrader
	
	// Subscription management
	subscriptions   map[string]map[*websocket.Conn]bool // shortCode -> connections
	subMutex        sync.RWMutex
	
	// Broadcasting
	broadcast       chan *models.RealtimeUpdate
	register        chan *websocket.Conn
	unregister      chan *websocket.Conn
	
	// Context for shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// RealtimeUpdate represents a real-time analytics update
type RealtimeUpdate struct {
	Type      string      `json:"type"`      // "click", "analytics_update", "conversion"
	ShortCode string      `json:"short_code"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewRealtimeAnalyticsService creates a new real-time analytics service
func NewRealtimeAnalyticsService(db *storage.PostgresStorage, cache *storage.RedisStorage, analyticsService *AnalyticsService) *RealtimeAnalyticsService {
	ctx, cancel := context.WithCancel(context.Background())
	
	service := &RealtimeAnalyticsService{
		db:              db,
		cache:           cache,
		analyticsService: analyticsService,
		clients:         make(map[*websocket.Conn]bool),
		subscriptions:   make(map[string]map[*websocket.Conn]bool),
		broadcast:       make(chan *models.RealtimeUpdate, 1000),
		register:        make(chan *websocket.Conn),
		unregister:      make(chan *websocket.Conn),
		ctx:             ctx,
		cancel:          cancel,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
	
	// Start the hub
	go service.hub()
	
	// Start periodic analytics updates
	go service.periodicUpdates()
	
	return service
}

// HandleWebSocket handles WebSocket connections for real-time analytics
func (r *RealtimeAnalyticsService) HandleWebSocket(w http.ResponseWriter, req *http.Request) {
	conn, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	
	// Register the new client
	r.register <- conn
	
	defer func() {
		r.unregister <- conn
		conn.Close()
	}()
	
	// Set connection timeouts
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	// Handle incoming messages (subscriptions)
	for {
		var msg models.WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		r.handleMessage(conn, &msg)
	}
}

// handleMessage processes incoming WebSocket messages
func (r *RealtimeAnalyticsService) handleMessage(conn *websocket.Conn, msg *models.WebSocketMessage) {
	switch msg.Type {
	case "subscribe":
		r.subscribe(conn, msg.ShortCode)
	case "unsubscribe":
		r.unsubscribe(conn, msg.ShortCode)
	case "ping":
		conn.WriteJSON(map[string]string{"type": "pong"})
	}
}

// subscribe adds a connection to the subscription list for a short code
func (r *RealtimeAnalyticsService) subscribe(conn *websocket.Conn, shortCode string) {
	r.subMutex.Lock()
	defer r.subMutex.Unlock()
	
	if r.subscriptions[shortCode] == nil {
		r.subscriptions[shortCode] = make(map[*websocket.Conn]bool)
	}
	r.subscriptions[shortCode][conn] = true
	
	// Send initial analytics data
	go r.sendInitialAnalytics(conn, shortCode)
}

// unsubscribe removes a connection from the subscription list for a short code
func (r *RealtimeAnalyticsService) unsubscribe(conn *websocket.Conn, shortCode string) {
	r.subMutex.Lock()
	defer r.subMutex.Unlock()
	
	if r.subscriptions[shortCode] != nil {
		delete(r.subscriptions[shortCode], conn)
		if len(r.subscriptions[shortCode]) == 0 {
			delete(r.subscriptions, shortCode)
		}
	}
}

// sendInitialAnalytics sends the current analytics data when a client subscribes
func (r *RealtimeAnalyticsService) sendInitialAnalytics(conn *websocket.Conn, shortCode string) {
	analytics, err := r.analyticsService.GetAnalytics(shortCode, 30) // Last 30 days
	if err != nil {
		return
	}
	
	update := &models.RealtimeUpdate{
		Type:      "initial_analytics",
		ShortCode: shortCode,
		Data:      analytics,
		Timestamp: time.Now(),
	}
	
	conn.WriteJSON(update)
}

// hub manages client connections and message broadcasting
func (r *RealtimeAnalyticsService) hub() {
	ticker := time.NewTicker(54 * time.Second) // Send ping every 54 seconds
	defer ticker.Stop()
	
	for {
		select {
		case conn := <-r.register:
			r.clientsMutex.Lock()
			r.clients[conn] = true
			r.clientsMutex.Unlock()
			log.Println("Client connected to real-time analytics")
			
		case conn := <-r.unregister:
			r.clientsMutex.Lock()
			if _, ok := r.clients[conn]; ok {
				delete(r.clients, conn)
				r.clientsMutex.Unlock()
				
				// Remove from all subscriptions
				r.subMutex.Lock()
				for shortCode, conns := range r.subscriptions {
					delete(conns, conn)
					if len(conns) == 0 {
						delete(r.subscriptions, shortCode)
					}
				}
				r.subMutex.Unlock()
				
				log.Println("Client disconnected from real-time analytics")
			} else {
				r.clientsMutex.Unlock()
			}
			
		case update := <-r.broadcast:
			r.broadcastUpdate(update)
			
		case <-ticker.C:
			// Send ping to all clients to keep connections alive
			r.clientsMutex.RLock()
			for conn := range r.clients {
				select {
				case <-time.After(time.Second):
					// Timeout writing to client, remove it
					go func(c *websocket.Conn) {
						r.unregister <- c
					}(conn)
				default:
					if err := conn.WriteJSON(map[string]string{"type": "ping"}); err != nil {
						go func(c *websocket.Conn) {
							r.unregister <- c
						}(conn)
					}
				}
			}
			r.clientsMutex.RUnlock()
			
		case <-r.ctx.Done():
			return
		}
	}
}

// broadcastUpdate sends an update to all subscribers of a short code
func (r *RealtimeAnalyticsService) broadcastUpdate(update *models.RealtimeUpdate) {
	r.subMutex.RLock()
	subscribers, exists := r.subscriptions[update.ShortCode]
	if !exists || len(subscribers) == 0 {
		r.subMutex.RUnlock()
		return
	}
	
	// Make a copy of subscribers to avoid holding the lock too long
	conns := make([]*websocket.Conn, 0, len(subscribers))
	for conn := range subscribers {
		conns = append(conns, conn)
	}
	r.subMutex.RUnlock()
	
	// Send to all subscribers
	for _, conn := range conns {
		select {
		case <-time.After(time.Second):
			// Timeout writing to client, remove it
			go func(c *websocket.Conn) {
				r.unregister <- c
			}(conn)
		default:
			if err := conn.WriteJSON(update); err != nil {
				go func(c *websocket.Conn) {
					r.unregister <- c
				}(conn)
			}
		}
	}
}

// BroadcastClick broadcasts a new click event in real-time
func (r *RealtimeAnalyticsService) BroadcastClick(shortCode, clientIP, userAgent, referrer string) {
	// Get geographic and device info for the click
	geoData := r.analyticsService.getGeographicData(clientIP)
	deviceData := r.analyticsService.getDeviceData(userAgent)
	
	clickData := models.RealtimeClickData{
		ShortCode: shortCode,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Referrer:  referrer,
		Country:   geoData.Country,
		City:      geoData.City,
		Device:    deviceData.DeviceType,
		Browser:   deviceData.Browser,
		OS:        deviceData.OS,
		Timestamp: time.Now(),
	}
	
	update := &models.RealtimeUpdate{
		Type:      "click",
		ShortCode: shortCode,
		Data:      clickData,
		Timestamp: time.Now(),
	}
	
	select {
	case r.broadcast <- update:
	case <-time.After(100 * time.Millisecond):
		// Drop the message if the broadcast channel is full
		log.Printf("Dropped real-time click update for %s (channel full)", shortCode)
	}
}

// BroadcastConversion broadcasts a conversion event in real-time
func (r *RealtimeAnalyticsService) BroadcastConversion(shortCode string, conversion *models.Conversion) {
	update := &models.RealtimeUpdate{
		Type:      "conversion",
		ShortCode: shortCode,
		Data:      conversion,
		Timestamp: time.Now(),
	}
	
	select {
	case r.broadcast <- update:
	case <-time.After(100 * time.Millisecond):
		log.Printf("Dropped real-time conversion update for %s (channel full)", shortCode)
	}
}

// periodicUpdates sends periodic analytics updates to subscribers
func (r *RealtimeAnalyticsService) periodicUpdates() {
	ticker := time.NewTicker(30 * time.Second) // Update every 30 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			r.sendPeriodicUpdates()
		case <-r.ctx.Done():
			return
		}
	}
}

// sendPeriodicUpdates sends updated analytics to all subscribed clients
func (r *RealtimeAnalyticsService) sendPeriodicUpdates() {
	r.subMutex.RLock()
	shortCodes := make([]string, 0, len(r.subscriptions))
	for shortCode := range r.subscriptions {
		shortCodes = append(shortCodes, shortCode)
	}
	r.subMutex.RUnlock()
	
	for _, shortCode := range shortCodes {
		go func(sc string) {
			analytics, err := r.analyticsService.GetAnalytics(sc, 1) // Last 24 hours for real-time
			if err != nil {
				return
			}
			
			update := &models.RealtimeUpdate{
				Type:      "analytics_update",
				ShortCode: sc,
				Data:      analytics,
				Timestamp: time.Now(),
			}
			
			select {
			case r.broadcast <- update:
			case <-time.After(100 * time.Millisecond):
				log.Printf("Dropped periodic analytics update for %s (channel full)", sc)
			}
		}(shortCode)
	}
}

// GetActiveSubscriptions returns the count of active subscriptions per short code
func (r *RealtimeAnalyticsService) GetActiveSubscriptions() map[string]int {
	r.subMutex.RLock()
	defer r.subMutex.RUnlock()
	
	result := make(map[string]int)
	for shortCode, conns := range r.subscriptions {
		result[shortCode] = len(conns)
	}
	return result
}

// GetActiveClients returns the total number of connected clients
func (r *RealtimeAnalyticsService) GetActiveClients() int {
	r.clientsMutex.RLock()
	defer r.clientsMutex.RUnlock()
	return len(r.clients)
}

// Shutdown gracefully shuts down the real-time analytics service
func (r *RealtimeAnalyticsService) Shutdown() {
	r.cancel()
	
	// Close all client connections
	r.clientsMutex.RLock()
	for conn := range r.clients {
		conn.Close()
	}
	r.clientsMutex.RUnlock()
}