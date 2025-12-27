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
	"fmt"
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
	Freq     string `json:"freq"`
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

type bandInfo struct {
	name string
	min  float64
	max  float64
}

var BAND_TABLE [33]bandInfo = [33]bandInfo{
	{"2190m", 0.1357, 0.1378},
	{"630m", 0.472, 0.479},
	{"560m", 0.501, 0.504},
	{"160m", 1.8, 2},
	{"80m", 3.5, 4},
	{"60m", 5.06, 5.45},
	{"40m", 7, 7.3},
	{"30m", 10.1, 10.15},
	{"20m", 14, 14.35},
	{"17m", 18.068, 18.168},
	{"15m", 21, 21.45},
	{"12m", 24.89, 24.99},
	{"10m", 28, 29.7},
	{"8m", 40, 45},
	{"6m", 50, 54},
	{"5m", 54.000001, 69.9},
	{"4m", 70, 71},
	{"2m", 144, 148},
	{"1.25m", 222, 225},
	{"70cm", 420, 450},
	{"33cm", 902, 928},
	{"23cm", 1240, 1300},
	{"13cm", 2300, 2450},
	{"9cm", 3300, 3500},
	{"6cm", 5650, 5925},
	{"3cm", 10000, 10500},
	{"1.25cm", 24000, 24250},
	{"6mm", 47000, 47200},
	{"4mm", 75500, 81000},
	{"2.5mm", 119980, 123000},
	{"2mm", 134000, 149000},
	{"1mm", 241000, 250000},
	{"submm", 300000, 7500000},
}

// tmpl & dict
var tmplJson []byte
var dictJson []byte

func write2adif(file *os.File, data LogLine) error {
	// 解析日期时间
	var dt [5]int
	var freq [2]float64
	fmt.Sscanf(data.Dt, "%d-%d-%dT%d:%d", &dt[0], &dt[1], &dt[2], &dt[3], &dt[4])
	fmt.Sscanf(data.Freq, "%f/%f", &freq[0], &freq[1])
	// 解析频段
	var bandTxIdx, bandRxIdx int
	for bandTxIdx = range BAND_TABLE {
		if freq[0]+freq[1] > BAND_TABLE[bandTxIdx].max {
			continue
		} else {
			break
		}
	}
	for bandRxIdx = range BAND_TABLE {
		if freq[0]+freq[1] > BAND_TABLE[bandRxIdx].max {
			continue
		} else {
			break
		}
	}
	// 解析频率
	var freqTx string = strconv.FormatFloat(freq[0]+freq[1], 'f', -1, 64)
	var freqRx string = strconv.FormatFloat(freq[0], 'f', -1, 64)
	// 格式化输出一行ADIF
	_, err := fmt.Fprintf(
		file,
		"<CALL:%d>%s <BAND:%d>%s <MODE:%d>%s <QSO_DATE:8>%04d%02d%02d <TIME_ON:6>%02d%02d00 <FREQ:%d>%s <BAND_RX:%d>%s <FREQ_RX:%d>%s <EOR>\n",
		len(data.Callsign), data.Callsign,
		len(BAND_TABLE[bandTxIdx].name), BAND_TABLE[bandTxIdx].name,
		len(data.Mode), data.Mode,
		dt[0], dt[1], dt[2],
		dt[3], dt[4],
		len(freqTx), freqTx,
		len(BAND_TABLE[bandRxIdx].name), BAND_TABLE[bandRxIdx].name,
		len(freqRx), freqRx,
	)
	// 检查错误
	if err != nil {
		return err
	}
	// 落盘
	return file.Sync()
}

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
	})

	// 启动服务
	log.Println("Server listening on local port 5973 ...")
	if err = http.ListenAndServe("localhost:5973", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
