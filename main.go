/*
	Ham Log Helper
	main server code
	by odorajbotoj (BG4QBF)
	2025.12
*/

package main

import (
	"embed"
	"encoding/csv"
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

const VERSION string = "v1.0.0"

//go:embed web/*
var embedFiles embed.FS

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type LogLine struct {
	Index    uint64 `json:"index"`
	Callsign string `json:"callsign"`
	Dt       string `json:"dt"`
	Band     string `json:"band"`
	Mode     string `json:"mode"`
	Rst      int    `json:"rst"`
	RRig     string `json:"rrig"`
	RPwr     string `json:"rpwr"`
	RAnt     string `json:"rant"`
	RQth     string `json:"rqth"`
	TRig     string `json:"trig"`
	TPwr     string `json:"tpwr"`
	TAnt     string `json:"tant"`
	TQth     string `json:"tqth"`
	Rmks     string `json:"rmks"`
}

// tmpl & dict
var tmplJson []byte
var dictJson []byte

func main() {
	log.Printf("\nHamLogHelper 业余无线电通联记录助手\nby odorajbotoj (BG4QBF)\nVERSION: %s", VERSION)

	// 读取天地图api-key
	tdtKeyBytes, err := os.ReadFile("tianditu-key.txt")
	tdtKey := strings.TrimSpace(string(tdtKeyBytes))
	if err != nil || tdtKey == "" {
		log.Fatalln("Cannot read tianditu-key.txt or file is empty.")
	}

	// 读取tmpl和dict
	if file, err := os.Open("tmpl.json"); err == nil {
		tmplJson, _ = io.ReadAll(file)
		file.Close()
	}
	if file, err := os.Open("dict.json"); err == nil {
		dictJson, _ = io.ReadAll(file)
		file.Close()
	}

	// 嵌入文件系统处理
	filesys, err := fs.Sub(embedFiles, "web")
	if err != nil {
		log.Fatalf("FileSystem Error: %v", err)
	}

	// 创建服务
	tmpl := template.Must(template.ParseFS(filesys, "index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.Execute(w, tdtKey); err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})
	fileServer := http.FileServer(http.FS(filesys))
	http.Handle("/css/", fileServer)
	http.Handle("/js/", fileServer)
	http.Handle("/favicon.ico", fileServer)
	http.HandleFunc("/tmpl.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write(tmplJson)
	})
	http.HandleFunc("/dict.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write(dictJson)
	})

	// 注册ws
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
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
					err = csvWriter.Write([]string{strconv.FormatUint(infoJson.Index, 10), infoJson.Callsign, infoJson.Dt, infoJson.Band, infoJson.Mode, strconv.Itoa(infoJson.Rst), infoJson.RRig, infoJson.RPwr, infoJson.RAnt, infoJson.RQth, infoJson.TRig, infoJson.TPwr, infoJson.TAnt, infoJson.TQth, infoJson.Rmks})
					if err != nil {
						log.Printf("CSV write failed: %v", err)
						return
					}
					csvWriter.Flush()
					if err = csvWriter.Error(); err != nil {
						log.Printf("CSV flush failed: %v", err)
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
	})

	// 启动服务
	log.Println("Server listening on local port 5973 ...")
	if err = http.ListenAndServe("localhost:5973", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
