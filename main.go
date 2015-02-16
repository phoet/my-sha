package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gopkg.in/gorp.v1"
	"net/http"
	"strconv"
	"time"
)

func routerHandlerFunc(router *mux.Router) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		router.ServeHTTP(res, req)
	}
}

func static(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "public"+req.URL.Path)
}

func homeResource(res http.ResponseWriter, req *http.Request) {
	id := getId(req)
	repo := findRepo(id)

	renderTemplate("index", repo, res)
}

func notFound(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "public/404.html")
}

var herokuAuth authenticator
var dbmap *gorp.DbMap

func revisionResource(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	var repo Repo
	err := dbmap.SelectOne(&repo, "select * from repos where token=$1", id)
	if err != nil {
		fmt.Println("did not find record: ", err, id)
		return
	}

	resp.Write([]byte(repo.Revision))
}

type hookResourceReq struct {
	App      string `json:"app"`
	User     string `json:"user"`
	Url      string `json:"url"`
	Head     string `json:"head"`
	PrevHead string `json:"prev_head"`
	HeadLong string `json:"head_long"`
	GitLog   string `json:"git_log"`
	Release  string `json:"release"`
}

func hookResource(resp http.ResponseWriter, req *http.Request) {
	id := getId(req)
	repo := findRepo(id)
	readForm(resp, req)
	fmt.Println("Form:", req.Form)
	reqD := buildRepo(req)
	repo.Revision = toJSON(reqD)
	update(repo)

	go notifyResource(repo)

	ok(resp)
}

func notifyResource(repo Repo) {
	user := mustGetenv("HEROKU_USERNAME")
	pass := mustGetenv("HEROKU_PASSWORD")

	url := "https://" + user + ":" + pass + "@api.heroku.com/vendor/apps/" + repo.App
	fmt.Println("URL:>", url)

	config := &notifyResourceResq{Config: buildConfig(repo)}
	jsonBody, err := json.Marshal(&config)
	check(err)
	fmt.Println("json", string(jsonBody))
	putJson(url, jsonBody)
}

type notifyResourceResq struct {
	Config map[string]string `json:"config"`
}

func buildRepo(req *http.Request) *hookResourceReq {
	hook := &hookResourceReq{
		App:      req.FormValue("app"),
		User:     req.FormValue("user"),
		Url:      req.FormValue("url"),
		Head:     req.FormValue("head"),
		PrevHead: req.FormValue("prev_head"),
		HeadLong: req.FormValue("head_long"),
		GitLog:   req.FormValue("git_log"),
		Release:  req.FormValue("release"),
	}
	fmt.Println("Hook:", hook)
	return hook
}

func ok(resp http.ResponseWriter) {
	resp.Write([]byte("OK!"))
}

func delete(repo Repo) {
	count, err := dbmap.Delete(&repo)
	fmt.Println("Rows updated:", count)
	checkErr(err, "Delete failed")
}

func update(repo Repo) {
	count, err := dbmap.Update(&repo)
	fmt.Println("Rows updated:", count)
	checkErr(err, "Update failed")
}

func toJSON(obj interface{}) string {
	jsonBody, err := json.Marshal(&obj)
	checkErr(err, "Marshal failed")
	return string(jsonBody)
}

func getId(req *http.Request) string {
	vars := mux.Vars(req)
	id := vars["id"]
	fmt.Println("request-id", id)
	return id
}

func findRepo(id string) Repo {
	var repo Repo
	err := dbmap.SelectOne(&repo, "select * from repos where token=$1", id)
	checkErr(err, "Select failed")
	return repo
}

type createResourceReq struct {
	HerokuId    string            `json:"heroku_id"`
	Plan        string            `json:"plan"`
	CallbackUrl string            `json:"callback_url"`
	Options     map[string]string `json:"options"`
}
type createResourceResp struct {
	Id      string            `json:"id"`
	Config  map[string]string `json:"config"`
	Message string            `json:"message"`
}

func createResource(resp http.ResponseWriter, req *http.Request) {
	if !ensureAuth(resp, req, herokuAuth) {
		return
	}
	reqD := &createResourceReq{}
	if !readJson(resp, req, reqD) {
		return
	}
	repo := newRepo(reqD.HerokuId)

	err := dbmap.Insert(&repo)
	checkErr(err, "Insert failed")

	respD := &createResourceResp{
		Id:      repo.Token,
		Config:  buildConfig(repo),
		Message: "All set up!"}
	writeJson(resp, respD)
}

func buildConfig(repo Repo) map[string]string {
	rootUrl := mustGetenv("ROOT_URL")
	return map[string]string{
		"MY_SHA_TOKEN":           repo.Token,
		"MY_SHA_REVISION":        repo.Revision,
		"MY_SHA_URL":             rootUrl + "resources/" + repo.Token,
		"MY_SHA_DEPLOY_HOOK_URL": rootUrl + "hook/" + repo.Token,
		"MY_SHA_REVISION_URL":    rootUrl + "revision/" + repo.Token,
	}
}

type updateResourceReq struct {
	HerokuId string `json:"heroku_id"`
	Plan     string `json:"plan"`
}
type updateResourceResp struct {
	Config  map[string]string `json:"config"`
	Message string            `json:"message"`
}

func updateResource(resp http.ResponseWriter, req *http.Request) {
	if !ensureAuth(resp, req, herokuAuth) {
		return
	}
	reqD := &updateResourceReq{}
	if !readJson(resp, req, reqD) {
		return
	}
	id := getId(req)
	repo := findRepo(id)
	respD := &updateResourceResp{
		Config:  buildConfig(repo),
		Message: "All updated!"}
	writeJson(resp, respD)
}

type destroyResourceResp struct {
	Message string `json:"message"`
}

func destroyResource(resp http.ResponseWriter, req *http.Request) {
	if !ensureAuth(resp, req, herokuAuth) {
		return
	}
	id := getId(req)
	repo := findRepo(id)
	delete(repo)

	respD := &destroyResourceResp{Message: "All torn down!"}
	writeJson(resp, &respD)
}

func createSession(resp http.ResponseWriter, req *http.Request) {
	readForm(resp, req)
	ssoSalt := mustGetenv("SSO_SALT")
	id := req.FormValue("id")
	timestamp := req.FormValue("timestamp")
	token := req.FormValue("token")
	navData := req.FormValue("nav-data")
	hash := sha1String(id + ":" + ssoSalt + ":" + timestamp)
	if hash != token {
		resp.WriteHeader(403)
		resp.Write([]byte("{message: \"Invalid token\"}"))
		return
	}
	timestampLimit := int(time.Now().Unix() - (2 * 60))
	timestampInt, err := strconv.Atoi(timestamp)
	if (err != nil) || (timestampInt < timestampLimit) {
		resp.WriteHeader(403)
		resp.Write([]byte("{message: \"Invalid timestamp\"}"))
		return
	}
	http.SetCookie(resp, &http.Cookie{
		Name:  "heroku-nav-data",
		Value: navData})
	http.Redirect(resp, req, "/"+id, 302)
}

func router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", static).Methods("GET")
	router.HandleFunc("/{id}", homeResource).Methods("GET")
	router.HandleFunc("/revision/{id}", revisionResource).Methods("GET")
	router.HandleFunc("/hook/{id}", hookResource).Methods("POST")
	router.HandleFunc("/heroku/resources", createResource).Methods("POST")
	router.HandleFunc("/heroku/resources/{id}", updateResource).Methods("PUT")
	router.HandleFunc("/heroku/resources/{id}", destroyResource).Methods("DELETE")
	router.HandleFunc("/sso/login", createSession).Methods("POST")
	router.NotFoundHandler = http.HandlerFunc(notFound)
	return router
}

func main() {
	dbmap = initDb()
	defer dbmap.Db.Close()

	logs := make(chan string, 10000)
	go runLogging(logs)

	herokuPassword := mustGetenv("HEROKU_PASSWORD")
	herokuAuth = func(u string, p string) bool {
		return p == herokuPassword
	}

	handler := routerHandlerFunc(router())
	handler = wrapLogging(handler, logs)

	port := mustGetenv("PORT")
	logs <- fmt.Sprintf("serve at=start port=%s", port)
	err := http.ListenAndServe(":"+port, handler)
	check(err)
}
