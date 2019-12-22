package main

const (
	Discord = iota
	IRC
)

// TODO: implement user listing
type Protocol interface {
	SendMessage(msg TextMessage)
	// TODO: EditMessage()

	JoinChannel(msg JoinChannelMessage)
	LeaveChannel(msg LeaveChannelMessage)

	Close()
}