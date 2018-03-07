package main

import (
	"net/http"
	"log"
	"io"
	"html/template"
	"os"
	"path/filepath"
	"flag"
	"github.com/aiganbao/go-swagger/asset"
)

const tpl = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta http-equiv="x-ua-compatible" content="IE=edge">
    <title>爱肝宝</title>
    <link rel="icon" type="image/png" href="swagger-ui/images/favicon-32x32.png" sizes="32x32"/>
    <link rel="icon" type="image/png" href="swagger-ui/images/favicon-16x16.png" sizes="16x16"/>
    <link href='swagger-ui/css/typography.css' media='screen' rel='stylesheet' type='text/css'/>
    <link href='swagger-ui/css/reset.css' media='screen' rel='stylesheet' type='text/css'/>
    <link href='swagger-ui/css/screen.css' media='screen' rel='stylesheet' type='text/css'/>
    <link href='swagger-ui/css/reset.css' media='print' rel='stylesheet' type='text/css'/>
    <link href='swagger-ui/css/print.css' media='print' rel='stylesheet' type='text/css'/>

    <script src='swagger-ui/lib/object-assign-pollyfill.js' type='text/javascript'></script>
    <script src='swagger-ui/lib/jquery-1.8.0.min.js' type='text/javascript'></script>
    <script src='swagger-ui/lib/jquery.slideto.min.js' type='text/javascript'></script>
    <script src='swagger-ui/lib/jquery.wiggle.min.js' type='text/javascript'></script>
    <script src='swagger-ui/lib/jquery.ba-bbq.min.js' type='text/javascript'></script>
    <script src='swagger-ui/lib/handlebars-4.0.5.js' type='text/javascript'></script>
    <script src='swagger-ui/lib/lodash.min.js' type='text/javascript'></script>
    <script src='swagger-ui/lib/backbone-min.js' type='text/javascript'></script>
    <script src='swagger-ui/swagger-ui.js' type='text/javascript'></script>
    <script src='swagger-ui/lib/highlight.9.1.0.pack.js' type='text/javascript'></script>
    <script src='swagger-ui/lib/highlight.9.1.0.pack_extended.js' type='text/javascript'></script>
    <script src='swagger-ui/lib/jsoneditor.min.js' type='text/javascript'></script>
    <script src='swagger-ui/lib/marked.js' type='text/javascript'></script>
    <script src='swagger-ui/lib/swagger-oauth.js' type='text/javascript'></script>

    <script src='swagger-ui/lang/translator.js' type='text/javascript'></script>
    <script src='swagger-ui/lang/zh-cn.js' type='text/javascript'></script>

    <script type="text/javascript">
        $(function () {

            // Pre load translate...
            if (window.SwaggerTranslator) {
                window.SwaggerTranslator.translate();
            }
            window.swaggerUi = new SwaggerUi({
                url: "/swagger/"+  {{.filename}},
                dom_id: "swagger-ui-container",
                supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch'],
                onComplete: function (swaggerApi, swaggerUi) {
                    if (typeof initOAuth == "function") {
                        initOAuth({
                            clientId: "your-client-id",
                            clientSecret: "your-client-secret-if-required",
                            realm: "your-realms",
                            appName: "your-app-name",
                            scopeSeparator: " ",
                            additionalQueryStringParams: {}
                        });
                    }

                    if (window.SwaggerTranslator) {
                        window.SwaggerTranslator.translate();
                    }
                },
                onFailure: function (data) {
                    log("Unable to Load SwaggerUI");
                },
                docExpansion: "none",
                jsonEditor: false,
                defaultModelRendering: 'schema',
                showRequestHeaders: false,
                showOperationIds: false
            });

            window.swaggerUi.load();

            function log() {
                if ('console' in window) {
                    console.log.apply(console, arguments);
                }
            }
        });
    </script>
</head>

<body class="swagger-section">
<div id='header'>
    <div class="swagger-ui-wrap">
        <a id="logo" href="http://www.aiganbao.com"><img class="logo__img" alt="swagger" height="30" width="30"
                                                         src="swagger-ui/images/logo_small.png"/><span
                class="logo__title">爱肝宝</span></a>
        <form id='api_selector'>

            <div id='auth_container'></div>
            <div class='input'><a id="explore" class="header__btn" href="#" data-sw-translate>Explore</a></div>
        </form>
    </div>
</div>

<div id="message-bar" class="swagger-ui-wrap" data-sw-translate>&nbsp;</div>
<div id="swagger-ui-container" class="swagger-ui-wrap"></div>
</body>
</html>
`

func renderhtml(filename string, out io.Writer) error {

	m := map[string]interface{}{
		"filename": filename,
	}
	return template.Must(template.New("markdown").Parse(tpl)).Execute(out, m)
}

func HandleFuncHttp(w http.ResponseWriter, r *http.Request) {
	defer func() {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")
	}()

}

func HandleHttp(w http.ResponseWriter, r *http.Request) {

	HandleFuncHttp(w, r)
	if r.URL.Path == "/" || r.URL.Path == "/index.html" {

		f, err := os.Open(filepath.Join(".", "/swagger-ui/index.html"))
		if err != nil {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		}
		defer f.Close()

		io.Copy(w, f)
	}

}

func handleServerSwagger(w http.ResponseWriter, r *http.Request) {

	HandleFuncHttp(w, r)

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

var addr = flag.String("a", ":8083", "请输入服务端地址")

func main() {

	go http.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(&asset.AssetFs)))
	go http.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger"))))

	go http.HandleFunc("/swagger.html", handleServerSwagger)

	go http.HandleFunc("/", HandleHttp)

	flag.Parse()

	log.Fatal(http.ListenAndServe(*addr, nil))

}
