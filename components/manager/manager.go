package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db/dbmodels"
	"github.com/khanhnguyen02311/EchoChat-WS/components/handler"
	"github.com/khanhnguyen02311/EchoChat-WS/components/manager/connection"
	"github.com/khanhnguyen02311/EchoChat-WS/components/manager/message"
	"github.com/khanhnguyen02311/EchoChat-WS/components/services/proto"
	"github.com/khanhnguyen02311/EchoChat-WS/components/services/servicemodels"
	conf "github.com/khanhnguyen02311/EchoChat-WS/configurations"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"sync"
	"time"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Accepting all requests
		},
	}
)

// A ConnectionManager handles all connections to the EchoChat server and their respective clients.
//
// It uses Echo to handle the HTTP requests/metrics/logging and custom Gorilla Websockets to handle the WebSocket connections.
// Each client ID can have multiple connections stored in a slice.
type ConnectionManager struct {
	MessageReceivedCounter *prometheus.CounterVec
	MessageSentCounter     *prometheus.CounterVec
	connectionsByID        map[int][]*connection.WSConnection
	connectionsByIDMutex   sync.RWMutex
	server                 *echo.Echo
	db                     *db.ScyllaDB
}

func NewConnectionManager(e *echo.Echo, db *db.ScyllaDB, msgSentCounter *prometheus.CounterVec, msgReceivedCounter *prometheus.CounterVec) *ConnectionManager {
	return &ConnectionManager{
		MessageSentCounter:     msgSentCounter,
		MessageReceivedCounter: msgReceivedCounter,
		connectionsByID:        make(map[int][]*connection.WSConnection),
		connectionsByIDMutex:   sync.RWMutex{},
		server:                 e,
		db:                     db,
	}
}

func (manager *ConnectionManager) _validateClient(token string) (int, string) {
	conn, err := grpc.Dial(conf.PROTO_GRPC_HOST+":"+conf.PROTO_GRPC_PORT, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return 0, ""
	}
	defer conn.Close()
	client := proto.NewEchoChatBEClient(conn)
	resp, err := client.ValidateToken(context.Background(), &proto.TokenValue{Token: token})
	if err != nil {
		return 0, ""
	}
	if resp.GetId() == -1 {
		fmt.Println("Invalid token: " + resp.GetName())
		return 0, ""
	}
	return int(resp.GetId()), resp.GetName()
}

func (manager *ConnectionManager) _getConnectionsMutex(clientID int) []*connection.WSConnection {
	manager.connectionsByIDMutex.RLock()
	defer manager.connectionsByIDMutex.RUnlock()
	return manager.connectionsByID[clientID]
}

func (manager *ConnectionManager) _addConnectionsMutex(conn *connection.WSConnection) {
	manager.connectionsByIDMutex.Lock()
	defer manager.connectionsByIDMutex.Unlock()
	manager.connectionsByID[conn.ClientID] = append(manager.connectionsByID[conn.ClientID], conn)
}

func (manager *ConnectionManager) _removeConnectionsMutex(conn *connection.WSConnection) {
	manager.connectionsByIDMutex.Lock()
	defer manager.connectionsByIDMutex.Unlock()
	// 1 account should have just 1 to some connections, so we should be fine to block the whole map
	for i, c := range manager.connectionsByID[conn.ClientID] {
		if conn != c {
			continue
		}
		if len(manager.connectionsByID[c.ClientID]) == 1 {
			delete(manager.connectionsByID, c.ClientID)
		} else {
			manager.connectionsByID[c.ClientID] = append(
				manager.connectionsByID[c.ClientID][:i], manager.connectionsByID[c.ClientID][i+1:]...)
		}
		break
	}
}

func (manager *ConnectionManager) _sendNotificationsForNewMessage(messageDB *dbmodels.Message, messageRMQ *servicemodels.RMQMessage) {
	notification := dbmodels.Notification{
		AccountinfoID: 0, // iterated later
		Type:          dbmodels.DBNotificationType[0],
	}
	// bind data to notification
	if messageDB != nil {
		notification.GroupID = messageDB.GroupID
		notification.AccountinfoIDSender = messageDB.AccountinfoID
		notification.Content = messageDB.AccountinfoName + ": " + messageDB.Content
		notification.TimeCreated = messageDB.TimeCreated.UTC()
	} else {
		notification.GroupID = messageRMQ.GroupID
		notification.AccountinfoIDSender = messageRMQ.AccountinfoID
		notification.Content = messageRMQ.AccountinfoName + ": " + messageRMQ.Content
		notification.TimeCreated = messageRMQ.TimeCreated.UTC()
	}
	// query all participants of that group
	listID, err := handler.NewParticipantHandler(manager.db).GetAllParticipantIDsFromGroup(notification.GroupID)
	if err != nil {
		fmt.Println("Error getting all participants:", err.Error())
		return
	}
	if len(listID) == 0 {
		fmt.Println("No participants found")
		return
	}
	// add notification to database
	notificationHandler := handler.NewNotificationHandler(manager.db)
	notificationHandler.AddMultipleNotifications(listID, &notification)
	// send notification to clients
	notificationMsg := message.NewOutputMessage(message.MsgTypeNotification, message.MsgStatusNew, "")
	notificationMsg.Notification = &notification
	manager.SendToClients(listID, notificationMsg)
}

func (manager *ConnectionManager) GetAllConnections() {
	fmt.Println("All connections:", manager.connectionsByID)
}

func (manager *ConnectionManager) ProcessInputMessage(conn *connection.WSConnection, msg *message.InputMessage) error {
	switch msg.Type {
	case message.MsgTypeMessageNew:
		manager.MessageReceivedCounter.WithLabelValues(message.MsgTypeMessageNew).Inc()
		groupHandler := handler.NewGroupHandler(manager.db)
		participantHandler := handler.NewParticipantHandler(manager.db)
		group, err := groupHandler.GetGroupByID(msg.Data.GroupID)
		if err != nil || group == nil {
			return errors.New("group not found")
		}
		participant, err := participantHandler.CheckJoinedParticipant(conn.ClientID, group.ID)
		if err != nil || participant == nil {
			return errors.New("not a participant of this group")
		}
		newMessage := dbmodels.Message{
			AccountinfoID:   conn.ClientID,
			GroupID:         msg.Data.GroupID,
			Content:         msg.Data.Content,
			TimeCreated:     time.Now().UTC(),
			Type:            msg.Data.Type,
			AccountinfoName: conn.ClientName,
			GroupName:       group.Name,
		}
		err = handler.NewMessageHandler(manager.db).AddNewMessage(&newMessage)
		if err != nil {
			return err
		}
		manager._sendNotificationsForNewMessage(&newMessage, nil)
		return nil

	case message.MsgTypeNotificationRead:
		manager.MessageReceivedCounter.WithLabelValues(message.MsgTypeNotificationRead).Inc()
		notificationSeen := dbmodels.NotificationSeen{
			AccountinfoID: conn.ClientID,
			Type:          msg.Data.Type,
			GroupID:       msg.Data.GroupID,
			TimeCreated:   time.Now().UTC(),
		}
		err := handler.NewNotificationHandler(manager.db).AddNotificationSeen(&notificationSeen)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("invalid message type")
}

func (manager *ConnectionManager) ProcessRMQMessage(msg []byte) {
	//fmt.Println("Received message from message queue:", string(msg))
	// run each process in a new coroutine
	go func() {
		parsedMsg := servicemodels.RMQMessage{}
		if err := json.Unmarshal(msg, &parsedMsg); err != nil {
			fmt.Println("Error unmarshalling message:", err.Error())
		}
		manager._sendNotificationsForNewMessage(nil, &parsedMsg)
	}()
}

func (manager *ConnectionManager) ProcessRMQNoti(msg []byte) {
	// fmt.Println("Received message from noti queue:", string(msg))
	// TODO: send noti to client
}

func (manager *ConnectionManager) AddConnection(conn *websocket.Conn, clientID int, clientName string) *connection.WSConnection {
	c := &connection.WSConnection{
		Conn:       conn,
		ClientID:   clientID,
		ClientName: clientName,
	}
	manager._addConnectionsMutex(c)
	return c
}

func (manager *ConnectionManager) RemoveConnection(c *connection.WSConnection) {
	defer c.Conn.Close()
	manager._removeConnectionsMutex(c)
}

func (manager *ConnectionManager) ValidateAndAddConnection(w http.ResponseWriter, r *http.Request, respHeader http.Header) (*connection.WSConnection, error) {
	token := r.URL.Query().Get("token")
	if token == "" {
		return nil, errors.New("missing token")
	}
	clientID, clientName := manager._validateClient(token)
	if clientID == 0 {
		return nil, errors.New("client not found")
	}
	ws, err := upgrader.Upgrade(w, r, respHeader)
	if err != nil {
		return nil, err
	}
	return manager.AddConnection(ws, clientID, clientName), nil
}

func (manager *ConnectionManager) SendToClient(clientID int, outputMessage *message.OutputMessage) {
	conns := manager._getConnectionsMutex(clientID)
	for _, conn := range conns {
		manager.MessageSentCounter.WithLabelValues(message.MsgTypeNotification).Inc()
		err := conn.WriteJSONMessage(outputMessage)
		if err != nil {
			fmt.Println("Error sending message to client:", err.Error())
		}
	}
}

func (manager *ConnectionManager) SendToClients(clientIDs []int, outputMessage *message.OutputMessage) {
	for _, clientID := range clientIDs {
		manager.SendToClient(clientID, outputMessage)
	}
}

func (manager *ConnectionManager) SendToAll(outputMessage *message.OutputMessage) {
	for clientID, _ := range manager.connectionsByID {
		manager.SendToClient(clientID, outputMessage)
	}
}

//func (manager *ConnectionManager) SendToClient(clientID int, outputMessage *message.OutputMessage) error {
//	wg := new(errgroup.Group)
//	for _, c := range manager.connectionsByID[clientID] {
//		c := c // https://golang.org/doc/faq#closures_and_goroutines
//		wg.Go(func() error {
//			fmt.Printf("Sending message to client %d at %s\n", c.ClientID, c.Conn.RemoteAddr())
//			return c.WriteJSONMessage(outputMessage)
//		})
//	}
//	// handle first non-nil error if needed
//	if err := wg.Wait(); err != nil {
//		return err
//	}
//	return nil
//}

//func (manager *ConnectionManager) SendToClients(clientIDs []int, outputMessage *message.OutputMessage) error {
//	wg := new(errgroup.Group)
//	for client := range clientIDs {
//		fmt.Println("Client ID:", client)
//		for _, c := range manager.connectionsByID[client] {
//			c := c // https://golang.org/doc/faq#closures_and_goroutines
//			wg.Go(func() error {
//				fmt.Printf("Sending message to client %d at %s\n", c.ClientID, c.Conn.RemoteAddr())
//				return c.WriteJSONMessage(outputMessage)
//			})
//		}
//	}
//	// handle first non-nil error if needed
//	if err := wg.Wait(); err != nil {
//		return err
//	}
//	return nil
//}
//
//func (manager *ConnectionManager) SendToAll(outputMessage *message.OutputMessage) error {
//	wg := new(errgroup.Group)
//	for _, conns := range manager.connectionsByID {
//		for _, c := range conns {
//			c := c // https://golang.org/doc/faq#closures_and_goroutines
//			wg.Go(func() error {
//				fmt.Printf("Sending message to client %d at %s\n", c.ClientID, c.Conn.RemoteAddr())
//				return c.WriteJSONMessage(outputMessage)
//			})
//		}
//	}
//	// handle first non-nil error if needed
//	if err := wg.Wait(); err != nil {
//		return err
//	}
//	return nil
//}
