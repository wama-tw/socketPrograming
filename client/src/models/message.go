package models

type Message struct {
	Author string `json:"author"`
	Text   string `json:"text"`
	Time   string `json:"time"`
	Exit   bool   `json:"exit"`
	Join   bool   `json:"join"`
}
