package autocontext

import (
	"fmt"
	"errors"
	"regexp"
	"net/http"
	"strings"
	"sync"
	"runtime"
)

var (
	contextMutex sync.Mutex
	contextData  = make(map[string]map[string]interface{})
	regexServeHTTP = regexp.MustCompile("\\.ServeHTTP\\(0x[a-fA-F0-9]+, 0x[a-fA-F0-9]+, 0x[a-fA-F0-9]+, (0x[a-fA-F0-9]+)\\)")
)

// Set value in context
func Set(r *http.Request, key string, val interface{}) {
	contextMutex.Lock()
	defer contextMutex.Unlock()
	addr := fmt.Sprintf("%p", r)
	if _, ok := contextData[addr]; !ok {
		contextData[addr] = make(map[string]interface{})
	}
	contextData[addr][key] = val
}

// Get value from context
func Get(r *http.Request, key string) interface{} {
	addr := fmt.Sprintf("%p", r)
	contextMutex.Lock()
	defer contextMutex.Unlock()
	if m, ok := contextData[addr]; ok {
		return m[key]
	}
	return nil
}

// GetAuto ruturns value from current *http.Request context. It is automatically extracted from stacktrace. 
func GetAuto(key string) interface{} {
	addr, err := getRequestAddress()
	if err != nil {
		return nil
	}
	contextMutex.Lock()
	defer contextMutex.Unlock()
	if m, ok := contextData[addr]; ok {
		return m[key]
	}
	return nil
}

// Clean function removes specific context busket from map
func Clean(r *http.Request) {
	contextMutex.Lock()
	defer contextMutex.Unlock()
	addr := fmt.Sprintf("%p", r)
	if _, ok := contextData[addr]; ok {
		delete(contextData, addr)
	}
}

// getRequestAddress this function extracts *http.Request address from the go-lang stacktrace string.
func getRequestAddress() (string, error) {
	trace := make([]byte, 2048)
	runtime.Stack(trace, false)
	//fmt.Printf("Stack of %d bytes: %s\n", count, trace)
	match := regexServeHTTP.FindStringSubmatch(trace)
	if len(match) != 2 {
		fmt.Println("Regex not found in stack")
		fmt.Printf("*** STACK ***\n%s\n", trace)
		return "", errors.New("Failed to find *http.Request address")
	}
	//fmt.Printf("*** extracted address from stacktrace: %s\n", match[1])
	return match[1], nil
}

