package cacheserver

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

type HTTPServer struct {
	LogFile string
}

// Listen starts the key-value server
func (server *HTTPServer) Listen(port string) {

	//create the cache
	cache := &Cache{
		LogFile: server.LogFile,
	}

	http.HandleFunc("/get/", func(w http.ResponseWriter, r *http.Request) {
		getHandler(w, r, cache)
	})

	http.HandleFunc("/set/", func(w http.ResponseWriter, r *http.Request) {
		setHandler(w, r, cache)
	})

	http.HandleFunc("/del/", func(w http.ResponseWriter, r *http.Request) {
		deleteHandler(w, r, cache)
	})

	http.HandleFunc("/commit", func(w http.ResponseWriter, r *http.Request) {
		getCommitIndex(w, r, cache)
	})

	http.HandleFunc("/logentry", func(w http.ResponseWriter, r *http.Request) {
		getLogEntries(w, r, cache)
	})

	http.HandleFunc("/appendentries/", func(w http.ResponseWriter, r *http.Request) {
		appendEntries(w, r, cache)
	})

	fmt.Println("key-value server listening on port", port)
	http.ListenAndServe(port, nil)
}

type StoreBody struct {
	Key   any `json:"key"`
	Value any `json:"value"`
}

func getCommitIndex(w http.ResponseWriter, r *http.Request, cache *Cache) {
	if r.Method == "GET" {
		fmt.Fprint(w, cache.commitIndex)
	} else {
		fmt.Fprint(w, "method not allowed\n")
	}
}

func getHandler(w http.ResponseWriter, r *http.Request, cache *Cache) {
	if r.Method == "GET" {
		path := strings.Split(r.URL.Path, "/")

		key := ""
		if len(path) >= 2 {
			key = path[2]
		}

		if key != "" {
			value := cache.getData(key)
			fmt.Fprint(w, value)

		} else {
			fmt.Fprint(w, "missing key\n")
		}
	} else {
		fmt.Fprint(w, "method not allowed\n")
	}
}

func setHandler(w http.ResponseWriter, r *http.Request, cache *Cache) {
	if r.Method == "POST" {
		path := strings.Split(r.URL.Path, "/")

		key := ""
		value := ""

		if len(path) >= 3 {
			key = path[2]
			value = path[3]
			cache.storeData(key, value)
			fmt.Fprint(w, "OK\n")
		} else {
			fmt.Fprint(w, "\n")
		}

	} else {
		fmt.Fprint(w, "method not allowed\n")
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request, cache *Cache) {
	if r.Method == "DELETE" {

		path := strings.Split(r.URL.Path, "/")

		key := ""
		if len(path) >= 2 {
			key = path[2]
		}

		if key != "" {
			value := cache.deleteData(key)
			fmt.Fprint(w, value)
		} else {
			fmt.Fprint(w, "missing key\n")
		}
	} else {
		fmt.Fprint(w, "method not allowed\n")
	}
}

func getLogEntries(w http.ResponseWriter, r *http.Request, cache *Cache) {
	if r.Method == "GET" {
		logEntry := strings.Join(cache.LogEntry, "")
		logEntryBase64 := base64.StdEncoding.EncodeToString([]byte(logEntry))
		fmt.Fprint(w, logEntryBase64)
	} else {
		fmt.Fprint(w, "method not allowed\n")
	}
}

func appendEntries(w http.ResponseWriter, r *http.Request, cache *Cache) {
	if r.Method == "GET" {
		path := strings.Split(r.URL.Path, "/")

		res := cache.appendEntries(path[2])
		fmt.Fprint(w, res)
	} else {
		fmt.Fprint(w, "method not allowed\n")
	}
}
