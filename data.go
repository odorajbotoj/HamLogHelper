package main

import "strconv"

type DataType int

const (
	ServerDefault DataType = iota
	ConnectData
	MessageData
	JsonData
)

type LogLine struct {
	Index    int64  `json:"index"`
	Callsign string `json:"callsign"`
	Dt       string `json:"dt"`
	Freq     string `json:"freq"`
	Mode     string `json:"mode"`
	Rst      any    `json:"rst"`
	RRig     string `json:"rrig"`
	RAnt     string `json:"rant"`
	RPwr     string `json:"rpwr"`
	RQth     string `json:"rqth"`
	TRig     string `json:"trig"`
	TAnt     string `json:"tant"`
	TPwr     string `json:"tpwr"`
	TQth     string `json:"tqth"`
	Rmks     string `json:"rmks"`
}

func (ll LogLine) getRst() string {
	switch v := ll.Rst.(type) {
	case float64:
		return strconv.Itoa(int(v))
	case string:
		return v
	default:
		return ""
	}
}

type Data struct {
	Type    DataType `json:"type"`
	Message string   `json:"message"`
	Payload LogLine  `json:"payload"`
}

type ExportData struct {
	Name string
	Data []LogLine
}
