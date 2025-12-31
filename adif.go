package main

import (
	"fmt"
	"os"
	"strconv"
)

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
