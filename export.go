package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

var fields []string = []string{
	"位次", "呼号", "日期时间", "频率", "模式",
	"信号", "对方设备", "对方功率", "对方天线", "对方台址",
	"己方设备", "己方功率", "己方天线", "己方台址", "备注",
}

func getCellIdx(row, col int64) string {
	colStr := []byte{}
	for col > 0 {
		col -= 1
		colStr = append(colStr, byte('A'+col%26))
		col /= 26
	}
	return string(colStr) + strconv.FormatInt(row, 10)
}

func getADIF(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Set("Content-Disposition", "attachment; filename="+name+".adi")
	w.Header().Set("Content-Type", "text/plain")
	for _, line := range lines {
		err = writeADIF(w, line)
		if err != nil {
			http.Error(w, "Write data error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func getCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Bad Method.", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("n")
	if name == "" {
		http.Error(w, "Bad args.", http.StatusBadRequest)
		return
	}
	selects := r.URL.Query().Get("s")
	if selects == "" {
		http.Error(w, "您未选择任何列\nYou didn't select any column.", http.StatusBadRequest)
		return
	}
	if len(selects) > 15 {
		http.Error(w, "Bad args.", http.StatusBadRequest)
		return
	}
	bjthhmm := r.URL.Query().Get("bjthhmm")
	if bjthhmm == "" || len(bjthhmm) != 1 {
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
	w.Header().Set("Content-Disposition", "attachment; filename="+name+".csv")
	w.Header().Set("Content-Type", "application/csv")
	cw := csv.NewWriter(w)
	head := []string{}
	for _, s := range selects {
		if s-'a' < 15 {
			head = append(head, fields[s-'a'])
		}
	}
	cw.Write(head)
	for _, line := range lines {
		cl := []string{}
		for _, s := range selects {
			switch s - 'a' {
			case 0:
				cl = append(cl, strconv.FormatInt(line.Index, 10))
			case 1:
				cl = append(cl, line.Callsign)
			case 2:
				if bjthhmm == "1" {
					t, err := time.Parse("2006-01-02T15:04", line.Dt)
					if err == nil {
						cl = append(cl, t.Add(8*time.Hour).Format("15:04"))
					} else {
						cl = append(cl, err.Error())
					}
				} else {
					cl = append(cl, line.Dt)
				}
			case 3:
				cl = append(cl, line.Freq)
			case 4:
				cl = append(cl, line.Mode)
			case 5:
				cl = append(cl, strconv.Itoa(line.Rst))
			case 6:
				cl = append(cl, line.RRig)
			case 7:
				cl = append(cl, line.RPwr)
			case 8:
				cl = append(cl, line.RAnt)
			case 9:
				cl = append(cl, line.RQth)
			case 10:
				cl = append(cl, line.TRig)
			case 11:
				cl = append(cl, line.TPwr)
			case 12:
				cl = append(cl, line.TAnt)
			case 13:
				cl = append(cl, line.TQth)
			case 14:
				cl = append(cl, line.Rmks)
			}
		}
		cw.Write(cl)
	}
	cw.Flush()
}

func getXLSX(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Bad Method.", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("n")
	if name == "" {
		http.Error(w, "Bad args.", http.StatusBadRequest)
		return
	}
	selects := r.URL.Query().Get("s")
	if selects == "" {
		http.Error(w, "您未选择任何列\nYou didn't select any column.", http.StatusBadRequest)
		return
	}
	if len(selects) > 15 {
		http.Error(w, "Bad args.", http.StatusBadRequest)
		return
	}
	bjthhmm := r.URL.Query().Get("bjthhmm")
	if bjthhmm == "" || len(bjthhmm) != 1 {
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
	w.Header().Set("Content-Disposition", "attachment; filename="+name+".xlsx")
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	xlsx := excelize.NewFile()
	defer func() {
		if err := xlsx.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	var idx int64 = 1
	for _, s := range selects {
		if s-'a' < 15 {
			xlsx.SetCellStr("Sheet1", getCellIdx(1, idx), fields[s-'a'])
			idx++
		}
	}
	for i, line := range lines {
		idx = 1
		for _, s := range selects {
			switch s - 'a' {
			case 0:
				xlsx.SetCellInt("Sheet1", getCellIdx(int64(i+2), idx), line.Index)
				idx++
			case 1:
				xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.Callsign)
				idx++
			case 2:
				if bjthhmm == "1" {
					t, err := time.Parse("2006-01-02T15:04", line.Dt)
					if err == nil {
						xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), t.Add(8*time.Hour).Format("15:04"))
					} else {
						xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), err.Error())
					}
				} else {
					xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.Dt)
				}
				idx++
			case 3:
				xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.Freq)
				idx++
			case 4:
				xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.Mode)
				idx++
			case 5:
				xlsx.SetCellInt("Sheet1", getCellIdx(int64(i+2), idx), int64(line.Rst))
				idx++
			case 6:
				xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.RRig)
				idx++
			case 7:
				xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.RPwr)
				idx++
			case 8:
				xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.RAnt)
				idx++
			case 9:
				xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.RQth)
				idx++
			case 10:
				xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.TRig)
				idx++
			case 11:
				xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.TPwr)
				idx++
			case 12:
				xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.TAnt)
				idx++
			case 13:
				xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.TQth)
				idx++
			case 14:
				xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.Rmks)
				idx++
			}
		}
	}
	if err := xlsx.Write(w); err != nil {
		fmt.Fprint(w, err.Error())
	}
}
