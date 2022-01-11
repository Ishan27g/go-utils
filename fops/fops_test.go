//package fops
//
//import (
//	"os"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func Test_Read(t *testing.T) {
//	const csvFile = "./file.csv"
//	defer func() {
//		os.Remove(csvFile)
//	}()
//	var ticks = 10
//	var entries []Entry
//	for i := 1; i <= ticks; i++ {
//		entries = append(entries, mockEntry(i))
//	}
//
//	offset := write(csvFile, 0, &entries)
//	assert.Equal(t, 470, int(offset))
//
//	entries = nil
//	for i := 1; i <= ticks; i++ {
//		entries = append(entries, mockEntry(i))
//	}
//
//	offsetAgain := write(csvFile, offset, &entries)
//	assert.Equal(t, 470*2, int(offsetAgain))
//
//	_, off := read(csvFile)
//	assert.Equal(t, off, offsetAgain)
//}
