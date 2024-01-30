package rabbitmq

import (
	"sync"
)

var notifyOpenConn, notifySetupDone []chan struct{}
var muxNotifyOpenConn, muxNotifySetup sync.Mutex = sync.Mutex{}, sync.Mutex{}

// NotifyOpenConnection registers a channel to be notified when the connection is open
func NotifyOpenConnection(notify chan struct{}) {
	muxNotifyOpenConn.Lock()
	defer muxNotifyOpenConn.Unlock()
	notifyOpenConn = append(notifyOpenConn, notify)
}

func notifyOpenConnections() {
	muxNotifyOpenConn.Lock()
	defer muxNotifyOpenConn.Unlock()
	for _, notify := range notifyOpenConn {
		close(notify)
	}
	notifyOpenConn = make([]chan struct{}, 0)
}

// NotifySetupDone registers a channel to be notified when the setup is done by the Setup function
func NotifySetupDone(notify chan struct{}) {
	muxNotifySetup.Lock()
	defer muxNotifySetup.Unlock()
	notifySetupDone = append(notifySetupDone, notify)
}

func notifySetupIsDone() {
	muxNotifySetup.Lock()
	defer muxNotifySetup.Unlock()
	for _, notify := range notifySetupDone {
		close(notify)
	}
	notifySetupDone = make([]chan struct{}, 0)
}
