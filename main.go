/*
	Ham Log Helper
	main server code
	by odorajbotoj (BG4QBF)
	2026.01
*/

package main

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

const VERSION string = "v1.0.0"

//go:embed web/*
var embedFiles embed.FS

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type bandInfo struct {
	name string
	min  float64
	max  float64
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
	http.HandleFunc("/ws", wsService)

	// 启动服务
	log.Println("Server listening on local port 5973 ...")
	if err = http.ListenAndServe("localhost:5973", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
