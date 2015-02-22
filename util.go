package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func check(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func pid() {
	pid := os.Getpid()
	err := ioutil.WriteFile(".pid", []byte(strconv.Itoa(pid)), 0644)
	check(err)
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing " + k)
	}
	return v
}

func sha1String(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func generateUUID() string {
	u4, err := uuid.NewV4()
	if err != nil {
		panic("error: " + err.Error())
	}
	return u4.String()
}
