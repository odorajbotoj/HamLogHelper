package main

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
	Rst      int    `json:"rst"`
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

type Data struct {
	Type    DataType `json:"type"`
	Message string   `json:"message"`
	Payload LogLine  `json:"payload,omitempty"`
}

type ExportData struct {
	Name string
	Data []LogLine
}
