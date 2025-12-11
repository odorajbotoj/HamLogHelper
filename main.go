/*
	Ham Log Helper
	main server code
	by odorajbotoj (BG4QBF)
	2025.12
*/

package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const VERSION string = "v1.0.0"

//go:embed web/*
var embedFiles embed.FS

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var count uint64 = 0

func main() {
	log.Printf("HamLogHelper 业余无线电通联记录助手\nby odorajbotoj (BG4QBF)\nVERSION: %s", VERSION)

	// 读取天地图api-key
	tdtKeyBytes, err := os.ReadFile("tianditu-key.txt")
	tdtKey := strings.TrimSpace(string(tdtKeyBytes))
	if err != nil || tdtKey == "" {
		log.Fatalln("Cannot read tianditu-key.txt or file is empty.")
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

	// 注册ws
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()
		file, err := os.OpenFile(time.Now().UTC().Format(time.RFC3339)+".csv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Printf("File writer failed: %v", err)
			return
		}
		defer file.Close()
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			if messageType == websocket.TextMessage {
				s := string(p)
				if s == "QSL?" {
					if err := conn.WriteMessage(messageType, fmt.Appendf(nil, "QSL.%d", atomic.LoadUint64(&count))); err != nil {
						log.Println(err)
						return
					}
				} else {
					_, err = fmt.Fprintf(file, "\"%d\",%s\n", atomic.AddUint64(&count, 1), s)
					if err != nil {
						log.Printf("File write failed: %v", err)
						return
					}
					err = file.Sync()
					if err != nil {
						log.Printf("File write failed: %v", err)
						return
					}
					if err := conn.WriteMessage(messageType, fmt.Appendf(nil, "ADDLOG>\"%d\",%s", atomic.LoadUint64(&count), s)); err != nil {
						log.Println(err)
						return
					}
				}
			}
		}
	})

	// 启动服务
	log.Println("Server listening on local port 5973 ...")
	if err = http.ListenAndServe(":5973", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
