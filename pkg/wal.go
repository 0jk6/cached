package wal

import (
	"log"
	"os"
)

func AppendLog(filename, data string) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	//write to the file
	_, err = f.WriteString(data)
	if err != nil {
		log.Fatal(err)
	}

	//flush the buffer
	err = f.Sync()
	if err != nil {
		log.Fatal(err)
	}
}