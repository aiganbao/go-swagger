package main

import (
	"net/http"
	"log"
	"io"
	"html/template"
	"os"
	"flag"
	"github.com/aiganbao/go-swagger/asset"
	"errors"
	"path/filepath"
	"strings"
	"encoding/json"
	"io/ioutil"
)

func renderhtml(filename string, out io.Writer) error {

	m := map[string]interface{}{
		"filename": filename,
	}

	bytes, err := asset.Asset("swagger-ui/home.html")
	if err != nil {
		return errors.New("no found home  template  html")
	}

	return template.Must(template.New("markdown").Parse(string(bytes))).Execute(out, m)
}

func hasSuffix(url string, prefix []string) bool {

	for _, p := range prefix {
		if strings.HasSuffix(url, p) {
			return true
		}
	}
	return false
}

func handleFuncHttp(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Access-Control-Allow-Origin", "*")

	if hasSuffix(r.URL.Path, []string{".jpg", ".css", ".png", ".png", ".js", ".gif"}) {
		w.Header().Add("Cache-Control", "max-age=604800, must-revalidate")
		w.Header().Add("Pragma", "public")

	} else {
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")
	}

}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	handleFuncHttp(w, r)
	if r.URL.Path == "/" || r.URL.Path == "/index.html" {

		bytes, err := asset.Asset("swagger-ui/index.html")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		}
		w.Write(bytes)
	} else {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}

func handleServerSwagger(w http.ResponseWriter, r *http.Request) {

	handleFuncHttp(w, r)

	var code = 200
	var err error
	defer func() {
		if err != nil {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.WriteHeader(code)
			io.WriteString(w, err.Error())
		}
	}()

	r.ParseForm()

	if len(r.Form["path"]) > 0 {
		err = renderhtml(r.Form["path"][0], w)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		}
	} else {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	if os.IsNotExist(err) {
		code = 404
	}
	return
}

func handleProxy(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleFuncHttp(w, r)
		if r.URL.Path == "/swagger-ui/index.html" || r.URL.Path == "/swagger-ui/" || r.URL.Path == "/swagger/index.html" || r.URL.Path == "/swagger/" {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		}

		h.ServeHTTP(w, r)
	}
}

type ApiReponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type MeTa struct {
	Name string `json:"name"`
	Href string `json:"href"`
}

var (
	SUCCESS      = ApiReponse{http.StatusOK, "上传的文件成功"}
	ERROR        = ApiReponse{http.StatusBadRequest, "上传的文件失败"}
	NO_PERMISSON = ApiReponse{http.StatusUnauthorized, "没有权限访问"}
	VALIDATION   = ApiReponse{http.StatusForbidden, "上传的文件非法"}
)

func upload(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	if !(r.Method == "POST" && strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data")) {
		io.WriteString(w, reponse(VALIDATION))
		return
	} else if *token != r.Header.Get("token") {
		io.WriteString(w, reponse(NO_PERMISSON))
		return
	}

	r.ParseMultipartForm(32 << 20)

	file, handler, err := r.FormFile("file")
	serviceName := r.FormValue("serviceName")

	if err != nil {
		io.WriteString(w, reponse(ERROR))
		return
	}
	defer file.Close()

	fileExt := filepath.Ext(handler.Filename)

	if checkFile(fileExt) {
		io.WriteString(w, reponse(VALIDATION))
		return
	}

	f, _ := os.OpenFile("./swagger/"+handler.Filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0660)

	_, err = io.Copy(f, file)

	if err != nil {
		io.WriteString(w, reponse(ERROR))
		return
	}

	meta, err := ioutil.ReadFile("./swagger/meta.json")
	if err != nil {
		io.WriteString(w, reponse(ERROR))
		return
	}
	var msg []MeTa

	if err := json.Unmarshal(meta, &msg); err == nil {

		nMeta := MeTa{
			Name: serviceName,
			Href: "/swagger.html?path=" + handler.Filename,
		}
		slice := removeSlice(msg, nMeta)

		m := append(slice, nMeta)
		bytes, _ := json.MarshalIndent(m, "", "\t")

		f, _ := os.OpenFile("./swagger/meta.json", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0660)

		_, err := io.WriteString(f, string(bytes))

		if err != nil {
			io.WriteString(f, string(meta))
			io.WriteString(w, reponse(ERROR))
			return
		}
	}

	io.WriteString(w, reponse(SUCCESS))

}

func removeSlice(m [] MeTa, r MeTa) [] MeTa {
	index := 0
	endIndex := len(m) - 1
	var result = make([]MeTa, 0)
	for k, v := range m {
		if v.Name == r.Name {
			result = append(result, m[index:k]...)
			index = k + 1
		} else if k == endIndex {
			result = append(result, m[index:endIndex+1]...)
		}
	}
	return result
}
func reponse(reponse ApiReponse) string {
	bytes, err := json.Marshal(reponse)
	if err != nil {
		return "{\"code\":500,\"msg\":\"上传的文件失败\"}"
	}
	return string(bytes)
}

func checkFile(name string) bool {

	ext := []string{".yml", ".yaml"}

	for _, v := range ext {
		if v == name {
			return false
		}
	}
	return true
}

var (
	addr  = flag.String("a", ":8083", "请输入服务端地址")
	token = flag.String("t", "9a8ecfd2f0a1ea11fc577e40", "请输入服务端token值")
)

func main() {

	flag.Parse()
	ADDR := os.Getenv("ADDR")

	if ADDR != "" {
		addr = &ADDR
	}

	TOKEN := os.Getenv("TOKEN")
	if TOKEN != "" {
		token = &TOKEN
	}

	log.Printf("Listening on %s  ", *addr)

	http.HandleFunc("/swagger-ui/", handleProxy(http.StripPrefix("/swagger-ui/", http.FileServer(&asset.AssetFs))))
	http.HandleFunc("/swagger/", handleProxy(http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger")))))

	http.HandleFunc("/swagger.html", handleServerSwagger)

	http.HandleFunc("/v1/swagger/upload", upload)

	http.HandleFunc("/", handleHttp)

	log.Fatal(http.ListenAndServe(*addr, nil))

}
