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

func handleFuncHttp(w http.ResponseWriter, r *http.Request) {
	defer func() {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")
	}()
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

var addr = flag.String("a", ":8083", "请输入服务端地址")

func main() {

	go http.HandleFunc("/swagger-ui/", handleProxy(http.StripPrefix("/swagger-ui/", http.FileServer(&asset.AssetFs))))
	go http.HandleFunc("/swagger/", handleProxy(http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger")))))

	go http.HandleFunc("/swagger.html", handleServerSwagger)

	go http.HandleFunc("/", handleHttp)

	flag.Parse()
	log.Printf("Listening on %s  ", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))

}
