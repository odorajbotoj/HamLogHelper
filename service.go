package main

import (
	"bufio"
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
	// 升级协议
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
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
				file, err = os.OpenFile(inData.Message+".hjl", os.O_CREATE|os.O_RDWR, 0644)
				if err != nil {
					log.Printf("Failed to open log file. %v\n", err)
					return
				}
				// 重置光标位置
				_, err = file.Seek(0, io.SeekStart)
				if err != nil {
					log.Printf("File Seek Error. %v\n", err)
					return
				}
				// 按行读取解析
				lst := newLinkedList()
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					var line LogLine
					if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
						log.Printf("Failed to unmarshal a line. %v\n", err)
						continue
					}
					lst.set(line.Index, &line)
				}
				lines := lst.dumpTill(lst.maxPos)
				// 清空文件
				err = file.Truncate(0)
				if err != nil {
					log.Printf("Failed to truncate file. %v", err)
					return
				}
				_, err = file.Seek(0, io.SeekStart)
				if err != nil {
					log.Printf("File Seek Error. %v\n", err)
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
					bLine, err := json.Marshal(line)
					if err != nil {
						log.Printf("Cannot marshal json line data. %v\n", err)
						return
					}
					if _, err := file.Write(append(bLine, '\n')); err != nil {
						log.Printf("Failed to write to file. %v\n", err)
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
				err = file.Sync()
				if err != nil {
					log.Printf("File sync failed: %v", err)
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
