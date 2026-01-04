package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsService(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Bad Method.", http.StatusMethodNotAllowed)
		return
	}
	// 升级协议
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Client connected.")
	defer conn.Close()
	// log计数
	var count uint64 = 0
	// 文件
	var file *os.File
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	// 禁止获取句柄前写入数据
	var connStat bool = false
	// 主逻辑
	for {
		// 接受消息
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// 处理消息
		if messageType == websocket.TextMessage {
			// 反序列化
			var inData Data
			err := json.Unmarshal(p, &inData)
			if err != nil {
				log.Printf("Failed to unmarshal input data. %v\n", err)
				continue
			}
			if inData.Type == ConnectData {
				// 客户端握手
				// 写回普通信息OK
				var reData Data
				reData.Type = MessageData
				reData.Message = "OK"
				b, err := json.Marshal(reData)
				if err != nil {
					log.Printf("Cannot marshal json data. %v\n", err)
					return
				}
				if err := conn.WriteMessage(messageType, b); err != nil {
					log.Printf("Failed to write to client. %v\n", err)
					return
				}
				// 打开文件
				file, err = os.OpenFile(inData.Message+".hjl", os.O_CREATE|os.O_RDWR|os.O_SYNC, 0644)
				if err != nil {
					log.Printf("Failed to open log file. %v\n", err)
					return
				}
				// 解析
				lines, err := parseLog(file)
				if err != nil {
					return
				}
				// 写回文件和前端
				reData.Type = JsonData
				reData.Message = "SYNC"
				for _, line := range lines {
					reData.Payload = line
					bReData, err := json.Marshal(reData)
					if err != nil {
						log.Printf("Cannot marshal json data. %v\n", err)
						return
					}
					if err := conn.WriteMessage(messageType, bReData); err != nil {
						log.Printf("Failed to write to client. %v\n", err)
						return
					}
					atomic.AddUint64(&count, 1)
				}
				connStat = true
			} else if inData.Type == JsonData {
				// 客户端发来信息
				if !connStat {
					continue
				}
				reMsg := "EDIT"
				if inData.Payload.Index == 0 {
					inData.Payload.Index = atomic.AddUint64(&count, 1)
					reMsg = "ADD"
				}
				// 序列化
				bLine, err := json.Marshal(inData.Payload)
				if err != nil {
					log.Printf("Failed to marshal json line data. %v\n", err)
					return
				}
				// 光标设置到末尾
				_, err = file.Seek(0, io.SeekEnd)
				if err != nil {
					log.Printf("File Seek Error. %v\n", err)
					return
				}
				// 写入文件
				_, err = file.Write(append(bLine, '\n'))
				if err != nil {
					log.Printf("File write failed: %v", err)
					return
				}
				// 写入前端
				var reData Data
				reData.Type = JsonData
				reData.Message = reMsg
				reData.Payload = inData.Payload
				bReData, err := json.Marshal(reData)
				if err != nil {
					log.Printf("Failed to marshal json. %v\n", err)
					return
				}
				if err := conn.WriteMessage(messageType, bReData); err != nil {
					log.Printf("Failed to write to client. %v", err)
					return
				}
			}
		}
	}
}

func exportService(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Bad Method.", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("n")
	if name == "" {
		http.Error(w, "Bad args.", http.StatusBadRequest)
		return
	}
	file, err := os.OpenFile(name+".hjl", os.O_RDWR|os.O_SYNC, 0644)
	if err != nil {
		http.Error(w, "File error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	lines, err := parseLog(file)
	if err != nil {
		http.Error(w, "Parse data error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := exportTmpl.Execute(w, ExportData{name, lines}); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
