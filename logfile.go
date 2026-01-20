package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
)

func parseLog(file *os.File) ([]LogLine, error) {
	// 重置光标位置
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		log.Printf("File Seek Error. %v\n", err)
		return nil, err
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
	// 清空文件
	err = file.Truncate(0)
	if err != nil {
		log.Printf("Failed to truncate file. %v", err)
		return nil, err
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		log.Printf("File Seek Error. %v\n", err)
		return nil, err
	}
	// 写回文件
	lines := lst.cleanAndDump()
	for _, line := range lines {
		bLine, err := json.Marshal(line)
		if err != nil {
			log.Printf("Cannot marshal json line data. %v\n", err)
			return nil, err
		}
		if _, err := file.Write(append(bLine, '\n')); err != nil {
			log.Printf("Failed to write to file. %v\n", err)
			return nil, err
		}
	}
	return lines, nil
}
