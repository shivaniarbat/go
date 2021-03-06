package pubnub

import (
	"sync"
)

//
type Listener struct {
	Status   chan *PNStatus
	Message  chan *PNMessage
	Presence chan *PNPresence
}

func NewListener() *Listener {
	return &Listener{
		Status:   make(chan *PNStatus),
		Message:  make(chan *PNMessage),
		Presence: make(chan *PNPresence),
	}
}

type ListenerManager struct {
	sync.RWMutex
	ctx          Context
	listeners    map[*Listener]bool
	exitListener chan bool
	pubnub       *PubNub
}

func newListenerManager(ctx Context, pn *PubNub) *ListenerManager {
	return &ListenerManager{
		listeners:    make(map[*Listener]bool, 2),
		ctx:          ctx,
		exitListener: make(chan bool),
		pubnub:       pn,
	}
}

func (m *ListenerManager) addListener(listener *Listener) {
	m.Lock()

	m.listeners[listener] = true
	m.Unlock()
}

func (m *ListenerManager) removeListener(listener *Listener) {
	delete(m.listeners, listener)
}

func (m *ListenerManager) removeAllListeners() {
	m.Lock()
	for l := range m.listeners {
		delete(m.listeners, l)
	}
	m.Unlock()
}

func (m *ListenerManager) announceStatus(status *PNStatus) {
	go func() {
		m.RLock()
		for l := range m.listeners {
			select {
			case <-m.exitListener:
				m.pubnub.Config.Log.Println("announceStatus exitListener")
				return
			case l.Status <- status:
			}
		}
		m.RUnlock()
	}()
}

func (m *ListenerManager) announceMessage(message *PNMessage) {
	go func() {
		m.RLock()
		for l := range m.listeners {
			select {
			case <-m.exitListener:
				m.pubnub.Config.Log.Println("announceMessage exitListener")
				return
			case l.Message <- message:
			}
		}
		m.RUnlock()
	}()
}

func (m *ListenerManager) announcePresence(presence *PNPresence) {
	m.RLock()

	for l := range m.listeners {
		select {
		case l.Presence <- presence:
		}
	}
	m.RUnlock()
}

//
type PNStatus struct {
	Category              StatusCategory
	Operation             OperationType
	ErrorData             error
	Error                 bool
	TLSEnabled            bool
	StatusCode            int
	UUID                  string
	AuthKey               string
	Origin                string
	ClientRequest         interface{} // Should be same for non-google environment
	AffectedChannels      []string
	AffectedChannelGroups []string
}

type PNMessage struct {
	Message           interface{}
	UserMetadata      interface{}
	SubscribedChannel string
	ActualChannel     string
	Channel           string
	Subscription      string
	Publisher         string
	Timetoken         int64
}

type PNPresence struct {
	Event             string
	UUID              string
	SubscribedChannel string
	ActualChannel     string
	Channel           string
	Subscription      string
	Occupancy         int
	Timetoken         int64
	Timestamp         int64
	UserMetadata      map[string]interface{}
	State             interface{}
	Join              []string
	Leave             []string
	Timeout           []string
	HereNowRefresh    bool
}
