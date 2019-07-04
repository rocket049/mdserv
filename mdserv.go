package main

import (
	"context"
	"flag"
	"log"
	"text/template"

	"bufio"
	"fmt"

	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/skratchdot/open-golang/open"

	md "github.com/russross/blackfriday"
)

var header string = `<html>
<head>
<meta http-equiv="content-type" content="text/html;charset=utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<link type="text/css" rel="stylesheet" href="/style.css"/>
<title>markdown</title>
<head>
<body>
<h1>markdown files</h1>
`
var tail string = `
</body>
</html>
`
var mdTmpl = `<html>
<head>
<meta http-equiv="content-type" content="text/html;charset=utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<link type="text/css" rel="stylesheet" href="/style.css"/>
<title>{{.title}}</title>
<head>
<body>
<h1>{{.title}}</h1>
{{.body}}
</body>
</html>`

type myHandler struct {
	CssPath [3]string
}

func (self *myHandler) init_path() {
	self.CssPath[0] = path.Join(os.Getenv("HOME"), ".config/mdserv/style.css")
}

func (self *myHandler) func_root(resp http.ResponseWriter, req *http.Request) {
	var f *os.File
	var er1 error
	var href string
	f, er1 = os.Open(".")
	if er1 != nil {
		resp.Write([]byte("err open dir"))
	} else {
		defer f.Close()
		nfs, err := f.Readdir(-1)
		if err != nil {
			resp.Write([]byte("err read dir"))
		} else {
			resp.Write([]byte(header))
			n := 1
			for i := 0; i < len(nfs); i++ {
				//fmt.Println( nfs[i].Name() )
				if strings.HasSuffix(strings.ToLower(nfs[i].Name()), ".md") {
					href = fmt.Sprintf("%d : <a href=\"/md/%s\">%s</a><br>\n", n, nfs[i].Name(), nfs[i].Name())
					resp.Write([]byte(href))
					n++
				}
			}
			resp.Write([]byte(tail))
		}
	}
}

func (self *myHandler) func_md(resp http.ResponseWriter, req *http.Request) {
	if strings.HasSuffix(strings.ToLower(req.URL.Path[4:]), ".md") {
		data := make(map[string]string)
		data["title"] = req.URL.Path[4:]
		file1, err := os.Open(req.URL.Path[4:])
		if err != nil {
			resp.Write([]byte("error open file " + req.URL.Path[4:]))
		}
		stat1, _ := file1.Stat()
		var buf []byte = make([]byte, stat1.Size())
		n, _ := file1.Read(buf)
		file1.Close()
		if int64(n) == stat1.Size() {
			//r := md.HtmlRenderer(md.HTML_SKIP_HTML, "", "")
			//opts := md.EXTENSION_FENCED_CODE | md.EXTENSION_TABLES
			//body := md.Markdown(buf, r, opts)
			body := md.Run(buf, md.WithExtensions(md.CommonExtensions))
			data["body"] = string(body)
		} else {
			resp.Write([]byte("error read file "))
		}
		t := template.New("")
		t.Parse(mdTmpl)
		t.Execute(resp, data)
	} else {
		//http.ServeFile(resp, req, req.URL.Path[4:])
		file1, err := os.Open(req.URL.Path[4:])
		if err != nil {
			resp.Write([]byte("error open file "))
			return
		}
		defer file1.Close()
		stat1, _ := file1.Stat()
		h1 := resp.Header()
		fields := strings.Split(req.URL.Path[4:], ".")
		h1.Set("Content-Type", mime.TypeByExtension("."+fields[len(fields)-1]))
		//fmt.Println(mime.TypeByExtension( "."+fields[len(fields)-1]))
		h1.Set("Content-Length", strconv.Itoa(int(stat1.Size())))
		buf1 := bufio.NewReader(file1)
		resp.WriteHeader(200)
		var n int64
		for {
			n, _ = buf1.WriteTo(resp)
			if n == 0 {
				break
			}
		}
	}
}

func (self *myHandler) func_style(resp http.ResponseWriter, req *http.Request) {
	var file1 *os.File
	var err error
	for i := 0; i < len(self.CssPath); i++ {
		file1, err = os.Open(self.CssPath[i])
		//log.Println(self.CssPath[i])
		if err == nil {
			break
		}
	}
	//log.Println(err)
	if err != nil {
		resp.WriteHeader(200)
		resp.Write(defaultCss)
		return
	}
	defer file1.Close()
	buf1 := bufio.NewReader(file1)
	resp.WriteHeader(200)
	var n int64
	for {
		n, _ = buf1.WriteTo(resp)
		if n == 0 {
			break
		}
	}
}

func (self *myHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if "/" == req.URL.Path {
		//default md文件列表
		self.func_root(resp, req)
	} else if strings.HasPrefix(req.URL.Path, "/md/") {
		self.func_md(resp, req)
	} else if req.URL.Path == "/style.css" {
		self.func_style(resp, req)
	}
}

func main() {
	var d = flag.String("d", ".", "work directory.default:'.'")
	flag.Parse()
	os.Chdir(*d)
	var serv *myHandler = new(myHandler)
	serv.init_path()
	fmt.Println("URL: http://localhost:8680")
	lang := os.Getenv("LANG")
	if strings.HasPrefix(lang, "zh_CN.") {
		fmt.Println("修改样式请自定义CSS：", serv.CssPath[0])
	} else {
		fmt.Println("User Define CSS：", serv.CssPath[0])
	}
	server := &http.Server{
		Addr:    "localhost:8680",
		Handler: serv,
	}
	//go http.ListenAndServe("localhost:8680", serv)
	go server.ListenAndServe()
	defer server.Shutdown(context.Background())
	open.Run("http://localhost:8680")

	WaitSig()
	log.Println("Shutdown HTTP Server.")
}
