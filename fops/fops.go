package fops

//
//import (
//	"bufio"
//	"fmt"
//	"os"
//	"strings"
//	"time"
//)
//
//func write(filePath string, offset int64, data *[]Entry) int64 {
//	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0755)
//	if err != nil {
//		return -1
//	}
//	defer func(f *os.File) {
//		_ = f.Close()
//	}(file)
//	var csv string
//	for _, entry := range *data {
//		csv += entry.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00") + "," +
//			entry.Data.(string) + "\n"
//	}
//	written, err := file.WriteAt([]byte(csv), offset)
//	if err != nil {
//		fmt.Println("what", err)
//	}
//	return offset + int64(written)
//}
//func read(filePath string) (*[]Entry, int64) {
//	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0755)
//	if err != nil {
//		return nil, 0
//	}
//	defer func(f *os.File) {
//		_ = f.Close()
//	}(file)
//	var lines []Entry
//	scanner := bufio.NewScanner(file)
//	scanner.Split(bufio.ScanLines)
//	for scanner.Scan() {
//		lineWords := strings.Split(scanner.Text(), ",")
//		if len(lineWords) == 0 || len(lineWords) != 2 {
//			fmt.Println("Bad len")
//			continue
//		}
//		time, err := time.Parse(time.RFC3339Nano, lineWords[0])
//		if err != nil {
//			fmt.Println(err)
//			continue
//		}
//		lines = append(lines, Entry{
//			CreatedAt: time,
//			Data:      lineWords[1],
//		})
//	}
//	f, _ := file.Stat()
//	return &lines, f.Size()
//}
