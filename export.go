package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/xuri/excelize/v2"
)

var fields []string = []string{
	"位次", "呼号", "日期时间", "频率", "模式",
	"信号", "对方设备", "对方功率", "对方天线", "对方台址",
	"己方设备", "己方功率", "己方天线", "己方台址", "备注",
}

func getCellIdx(row, col int64) string {
	colStr := ""
	for col > 0 {
		col -= 1
		colStr += string('A' + col%26)
		col /= 26
	}
	return colStr + strconv.FormatInt(row, 10)
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
	file, err := os.OpenFile(name+".hjl", os.O_CREATE|os.O_RDWR, 0644)
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
	if selects == "" || len(selects) != 15 {
		http.Error(w, "Bad args.", http.StatusBadRequest)
		return
	}
	file, err := os.OpenFile(name+".hjl", os.O_CREATE|os.O_RDWR, 0644)
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
	for i, f := range fields {
		if selects[i] == '1' {
			head = append(head, f)
		}
	}
	cw.Write(head)
	for _, line := range lines {
		cl := []string{}
		if selects[0] == '1' {
			cl = append(cl, strconv.FormatUint(line.Index, 10))
		}
		if selects[1] == '1' {
			cl = append(cl, line.Callsign)
		}
		if selects[2] == '1' {
			cl = append(cl, line.Dt)
		}
		if selects[3] == '1' {
			cl = append(cl, line.Freq)
		}
		if selects[4] == '1' {
			cl = append(cl, line.Mode)
		}
		if selects[5] == '1' {
			cl = append(cl, strconv.Itoa(line.Rst))
		}
		if selects[6] == '1' {
			cl = append(cl, line.RRig)
		}
		if selects[7] == '1' {
			cl = append(cl, line.RPwr)
		}
		if selects[8] == '1' {
			cl = append(cl, line.RAnt)
		}
		if selects[9] == '1' {
			cl = append(cl, line.RQth)
		}
		if selects[10] == '1' {
			cl = append(cl, line.TRig)
		}
		if selects[11] == '1' {
			cl = append(cl, line.TPwr)
		}
		if selects[12] == '1' {
			cl = append(cl, line.TAnt)
		}
		if selects[13] == '1' {
			cl = append(cl, line.TQth)
		}
		if selects[14] == '1' {
			cl = append(cl, line.Rmks)
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
	if selects == "" || len(selects) != 15 {
		http.Error(w, "Bad args.", http.StatusBadRequest)
		return
	}
	file, err := os.OpenFile(name+".hjl", os.O_CREATE|os.O_RDWR, 0644)
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
	for i, f := range fields {
		if selects[i] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(1, idx), f)
			idx++
		}
	}
	for i, line := range lines {
		idx = 1
		if selects[0] == '1' {
			xlsx.SetCellUint("Sheet1", getCellIdx(int64(i+2), idx), line.Index)
			idx++
		}
		if selects[1] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.Callsign)
			idx++
		}
		if selects[2] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.Dt)
			idx++
		}
		if selects[3] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.Freq)
			idx++
		}
		if selects[4] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.Mode)
			idx++
		}
		if selects[5] == '1' {
			xlsx.SetCellInt("Sheet1", getCellIdx(int64(i+2), idx), int64(line.Rst))
			idx++
		}
		if selects[6] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.RRig)
			idx++
		}
		if selects[7] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.RPwr)
			idx++
		}
		if selects[8] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.RAnt)
			idx++
		}
		if selects[9] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.RQth)
			idx++
		}
		if selects[10] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.TRig)
			idx++
		}
		if selects[11] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.TPwr)
			idx++
		}
		if selects[12] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.TAnt)
			idx++
		}
		if selects[13] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.TQth)
			idx++
		}
		if selects[14] == '1' {
			xlsx.SetCellStr("Sheet1", getCellIdx(int64(i+2), idx), line.Rmks)
			idx++
		}
	}
	if err := xlsx.Write(w); err != nil {
		fmt.Fprint(w, err.Error())
	}
}
