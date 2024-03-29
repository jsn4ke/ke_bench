package kebench

type BenchMessage struct {
	Msg string
}

func ProcessMessage(msg *BenchMessage) *BenchMessage {
	return &BenchMessage{Msg: msg.Msg}
}
