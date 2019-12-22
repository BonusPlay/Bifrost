package main

type Channel struct {
	Name string
	Public bool
}

// TODO: more message types (reaction, attachment, etc)
type TextMessage struct {
	FromProtocol uint
	ToProtocol uint
	From string
	Text string
	Channel Channel
}

type JoinChannelMessage struct {
	FromProtocol uint
	ToProtocol uint
	Channel Channel
}

type LeaveChannelMessage struct {
	FromProtocol uint
	ToProtocol uint
	Channel Channel
}