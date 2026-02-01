/*
	Ham Log Helper
	main server code
	by odorajbotoj (BG4QBF)
	2026.01
*/

package main

import (
	"embed"
	"encoding/json"
	"flag"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const VERSION string = "v1.6.0"

//go:embed web/*
var embedFiles embed.FS

var indexTmpl, exportTmpl *template.Template

var setAddr = flag.String("a", "127.0.0.1:5973", "Server bind addr 服务绑定地址端口")
var silent = flag.Bool("s", false, "Silent start 启动时不自动打开浏览器")

// tmpl & dict
var tmplJson []byte
var dictJson []byte

func main() {
	log.Printf("\nHamLogHelper 业余无线电台网点名记录助手\nAuthor 作者: odorajbotoj (BG4QBF)\nVersion 版本: %s", VERSION)
	flag.Parse()

	// 读取天地图api-key
	tdtKeyBytes, _ := os.ReadFile("tianditu-key.txt")
	tdtKey := strings.TrimSpace(string(tdtKeyBytes))

	// 读取tmpl和dict
	if file, err := os.Open("tmpl.json"); err == nil {
		tmplJson, _ = io.ReadAll(file)
		file.Close()
	}
	if file, err := os.Open("dict.json"); err == nil {
		dictJson, _ = io.ReadAll(file)
		file.Close()
	}
	if !json.Valid(tmplJson) {
		tmplJson = []byte("[]")
	}
	if !json.Valid(dictJson) {
		dictJson = []byte("{ \"rig\": {}, \"ant\": {}, \"pwr\": {}, \"qth\": {} }")
	}

	// 嵌入文件系统处理
	filesys, err := fs.Sub(embedFiles, "web")
	if err != nil {
		log.Fatalf("FileSystem Error: %v", err)
	}

	// 创建服务
	indexTmpl = template.Must(template.ParseFS(filesys, "index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := indexTmpl.Execute(w, struct{ Key, Version string }{tdtKey, VERSION}); err != nil {
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

	// 导出功能
	exportTmpl = template.Must(template.ParseFS(filesys, "export.html"))
	http.HandleFunc("/export", exportService)
	http.HandleFunc("/getadif", getADIF)
	http.HandleFunc("/getcsv", getCSV)
	http.HandleFunc("/getxlsx", getXLSX)

	// 编辑数据库
	http.Handle("/dbeditor.html", fileServer)
	http.HandleFunc("/editdb", editdbService)

	go func() {
		resp, err := http.Get("https://api.github.com/repos/odorajbotoj/HamLogHelper/releases/latest")
		if err != nil {
			log.Printf("Update Checker 检查更新: ERROR - %v\n", err)
			return
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Update Checker 检查更新: ERROR - %v\n", err)
			return
		}
		newVer := struct {
			TagName string `json:"tag_name"`
		}{}
		json.Unmarshal(b, &newVer)
		if newVer.TagName > VERSION {
			log.Printf("Update Checker 检查更新: New version 新版本 - %s\nhttps://github.com/odorajbotoj/HamLogHelper/releases/latest", newVer.TagName)
		}
	}()

	// 启动服务
	listener, err := net.Listen("tcp", *setAddr)
	if err != nil {
		log.Fatalf("Listen Failed 服务器监听失败: %v", err)
	}
	defer listener.Close()
	bindAddr := listener.Addr().String()
	log.Printf("Server Starting 服务启动 ...\nURL 浏览器访问 http://%s 打开界面", bindAddr)

	if !*silent {
		go func() {
			time.Sleep(1 * time.Second) // delay and wait server starting
			switch runtime.GOOS {
			case "windows":
				exec.Command("cmd", "/c", "start", "http://"+bindAddr).Start()
			case "linux":
				exec.Command("xdg-open", "http://"+bindAddr).Start()
			case "darwin":
				exec.Command("open", "http://"+bindAddr).Start()
			default:
				log.Println("Browser Opener Error 无法自动打开浏览器")
			}
		}()
	}

	if err = http.Serve(listener, nil); err != nil {
		log.Fatalf("Server Failed 服务器运行失败: %v", err)
	}
}
