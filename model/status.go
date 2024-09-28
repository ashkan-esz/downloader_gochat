package model

import "sync"

type Status struct {
	Tasks *Tasks `json:"tasks"`
}

type Tasks struct {
	BlurHash         *TaskInfo `json:"blurHash"`
	MediaService     *TaskInfo `json:"mediaService"`
	Notification     *TaskInfo `json:"notification"`
	PushNotification *TaskInfo `json:"pushNotification"`
	TelegramMessage  *TaskInfo `json:"telegramMessage"`
}

type TaskInfo struct {
	ConsumerCount int         `json:"consumerCount"`
	Links         []string    `json:"links"`
	Mux           *sync.Mutex `json:"-"`
	RunningCount  int         `json:"runningCount"`
}
