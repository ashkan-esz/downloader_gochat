package model

type ClientMessage struct {
	Content    string `json:"content"`
	RoomId     string `json:"roomId"`
	ReceiverId string `json:"receiverId"`
}

type CreateRoomReq struct {
	SenderId   string `json:"senderId"`
	ReceiverId string `json:"receiverId"`
}

type CreateRoomRes struct {
	RoomId string `json:"roomId"`
}

type RoomRes struct {
	ID string `json:"id"`
}

type ClientRes struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
