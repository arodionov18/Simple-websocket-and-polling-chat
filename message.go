package main


type RequestType int

var (
	MSG_SEND = RequestType(1)
	MSG_GET = RequestType(2)
)

type Request struct {
	Type RequestType `json:"type"`
	Name *string     `json:"name"`
	Text *string	`json:"text"`
}


func NewSendRequest(name, text string) Request {
	return Request{
		Type: MSG_SEND,
		Name: &name,
		Text: &text,
	}
}

func NewGetRequest() Request {
	return Request{
		Type: MSG_GET,
	}
}

type Reply struct {
	Ok bool `json:"ok"`
	Error    *string   `json:"error"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Name string `json:"name"`
	Text string `json:"text"`
}