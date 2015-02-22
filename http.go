package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"
)

func markDowner(filename string) template.HTML {
	input, err := ioutil.ReadFile(filename)
	check(err)
	s := blackfriday.MarkdownCommon(input)
	return template.HTML(s)
}

func renderTemplate(view string, obj interface{}, w http.ResponseWriter) {
	tmpl, err := template.New("layout.html").Funcs(template.FuncMap{"markDown": markDowner}).ParseFiles(
		path.Join("templates", "layout.html"),
		path.Join("templates", "includes", "heroku.html"),
		path.Join("templates", "views", view+".html"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, obj); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type statusCapturingResponseWriter struct {
	status int
	http.ResponseWriter
}

func (w statusCapturingResponseWriter) WriteHeader(s int) {
	w.status = s
	w.ResponseWriter.WriteHeader(s)
}

func runLogging(logs chan string) {
	for log := range logs {
		fmt.Println(log)
	}
}

func wrapLogging(f http.HandlerFunc, logs chan string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		wres := statusCapturingResponseWriter{-1, res}
		start := time.Now()
		f(wres, req)
		method := req.Method
		path := req.URL.Path
		elapsed := float64(time.Since(start)) / 1000000.0
		logs <- fmt.Sprintf("request at=finish method=%s path=%s status=%d elapsed=%f",
			method, path, wres.status, elapsed)
	}
}

type authenticator func(string, string) bool

func testAuth(r *http.Request, auth authenticator) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 || s[0] != "Basic" {
		return false
	}
	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}
	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}
	return auth(pair[0], pair[1])
}

func denyAuth(res http.ResponseWriter) {
	res.Header().Set("WWW-Authenticate", `Basic realm="private"`)
	res.WriteHeader(401)
	res.Write([]byte("{message: \"Unauthorized\"}\n"))
}

func ensureAuth(res http.ResponseWriter, req *http.Request, auth authenticator) bool {
	if testAuth(req, auth) {
		return true
	}
	denyAuth(res)
	return false
}

func readForm(resp http.ResponseWriter, req *http.Request) bool {
	err := req.ParseForm()
	if err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte("{message: \"Invalid body\"}"))
		return false
	}
	return true
}

func readJson(resp http.ResponseWriter, req *http.Request, reqD interface{}) bool {
	err := json.NewDecoder(req.Body).Decode(reqD)
	if err != nil {
		fmt.Println("could not parse json", err.Error())
		resp.WriteHeader(400)
		resp.Write([]byte("{message: \"Invalid body\"}"))
		return false
	}
	return true
}

func writeJson(resp http.ResponseWriter, respD interface{}) {
	b, err := json.Marshal(&respD)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("{message: \"Internal server error\"}"))
	} else {
		resp.Write(b)
	}
}

func putJson(url string, jsonBody []byte) {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	check(err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	check(err)
	fmt.Println("response Body:", string(body))
}
