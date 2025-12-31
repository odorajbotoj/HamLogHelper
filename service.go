package main

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	var adif *os.File
	defer func() {
		if file != nil {
			file.Close()
		}
		if adif != nil {
			adif.Close()
		}
	}()
	// 禁止获取句柄前写入数据
	var connstat bool = false
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
			s := string(p)
			if strings.HasPrefix(s, "QSL?") { // 建立连接
				// 确认连接
				if err := conn.WriteMessage(messageType, []byte("QSL.")); err != nil {
					log.Println(err)
					return
				}
				// 取文件名
				fname := strings.TrimPrefix(s, "QSL?")
				// 打开文件
				if rfile, err := os.Open(fname + ".csv"); err == nil {
					// 读取csv
					csvReader := csv.NewReader(rfile)
					for {
						record, err := csvReader.Read()
						if err != nil {
							break
						}
						idx, err := strconv.ParseUint(record[0], 10, 64)
						if err != nil {
							log.Println(err)
							continue
						}
						atomic.StoreUint64(&count, idx)
						rst, err := strconv.Atoi(record[5])
						if err != nil {
							log.Println(err)
							continue
						}
						ll := LogLine{idx, record[1], record[2], record[3], record[4], rst, record[6], record[7], record[8], record[9], record[10], record[11], record[12], record[13], record[14]}
						infoJson, err := json.Marshal(ll)
						if err != nil {
							log.Println(err)
							continue
						}
						// 传输给前端
						if err := conn.WriteMessage(messageType, infoJson); err != nil {
							log.Println(err)
							return
						}
					}
					rfile.Close()
				}
				if file != nil {
					file.Close()
				}
				file, err = os.OpenFile(fname+".csv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
				if err != nil {
					log.Printf("File writer failed: %v", err)
					return
				}
				adif, err = os.OpenFile(fname+".adi", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
				if err != nil {
					log.Printf("ADIF writer failed: %v", err)
					return
				}
				connstat = true
			} else { // 传输log
				if !connstat {
					continue
				}
				var infoJson LogLine
				// 反序列化
				err := json.Unmarshal(p, &infoJson)
				if err != nil {
					log.Println(err)
					continue
				}
				// 序号自增
				infoJson.Index = atomic.AddUint64(&count, 1)
				// 序列化
				sending, err := json.Marshal(infoJson)
				if err != nil {
					log.Println(err)
					continue
				}
				// 写入csv
				csvWriter := csv.NewWriter(file)
				err = csvWriter.Write([]string{strconv.FormatUint(infoJson.Index, 10), infoJson.Callsign, infoJson.Dt, infoJson.Freq, infoJson.Mode, strconv.Itoa(infoJson.Rst), infoJson.RRig, infoJson.RPwr, infoJson.RAnt, infoJson.RQth, infoJson.TRig, infoJson.TPwr, infoJson.TAnt, infoJson.TQth, infoJson.Rmks})
				if err != nil {
					log.Printf("CSV write failed: %v", err)
					return
				}
				csvWriter.Flush()
				if err = csvWriter.Error(); err != nil {
					log.Printf("CSV flush failed: %v", err)
					return
				}
				// 写入adif
				if err = write2adif(adif, infoJson); err != nil {
					log.Printf("ADIF flush failed: %v", err)
					return
				}
				// 写回客户端
				if err := conn.WriteMessage(messageType, sending); err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}
