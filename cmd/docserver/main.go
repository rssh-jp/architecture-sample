// Package main はGoドキュメントをHTTPで配信するローカルサーバーです。
package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"path"
	"strings"
)

type packageDoc struct {
	Path string
	URL  string
	Text template.HTML
}

var pageTemplate = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="ja">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Goドキュメント - {{.Title}}</title>
	<style>
		body { margin: 0; color: #202124; font-family: sans-serif; }
		header { padding: 16px 24px; background: #087f8c; color: white; }
		main { display: grid; grid-template-columns: 260px 1fr; min-height: calc(100vh - 65px); }
		nav { padding: 24px; background: #f3f6f7; border-right: 1px solid #d8e0e2; }
		nav a { display: block; padding: 8px 0; color: #075e68; text-decoration: none; }
		article { padding: 24px 32px; max-width: 1000px; }
		pre { padding: 20px; overflow: auto; background: #f6f8f8; border: 1px solid #d8e0e2; line-height: 1.5; }
		@media (max-width: 700px) { main { display: block; } nav { border-right: 0; border-bottom: 1px solid #d8e0e2; } article { padding: 20px 16px; } }
	</style>
</head>
<body>
<header><strong>Goドキュメント</strong></header>
<main>
<nav>
	<strong>パッケージ</strong>
	{{range .Packages}}<a href="{{.URL}}">{{.Path}}</a>{{end}}
</nav>
<article>
	<h1>{{.Title}}</h1>
	{{.Content}}
</article>
</main>
</body>
</html>`))

type pageData struct {
	Title    string
	Packages []packageDoc
	Content  template.HTML
}

func main() {
	address := flag.String("addr", "localhost:7777", "HTTPサーバーの待ち受けアドレス")
	flag.Parse()

	packages, err := listPackages()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		packagePath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if packagePath == "" {
			render(w, pageData{Title: "パッケージ一覧", Packages: packages, Content: "<p>パッケージを選択してください。</p>"})
			return
		}

		for _, packageDoc := range packages {
			if packageDoc.Path == packagePath {
				output, err := goDoc(packagePath)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				render(w, pageData{Title: packagePath, Packages: packages, Content: template.HTML("<pre>" + template.HTMLEscapeString(output) + "</pre>")})
				return
			}
		}

		http.NotFound(w, r)
	})

	log.Printf("ドキュメントを http://%s/ で公開します", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}

func listPackages() ([]packageDoc, error) {
	output, err := exec.Command("go", "list", "./...").Output()
	if err != nil {
		return nil, fmt.Errorf("パッケージ一覧の取得に失敗しました: %w", err)
	}

	var packages []packageDoc
	for _, packagePath := range strings.Fields(string(output)) {
		packages = append(packages, packageDoc{Path: packagePath, URL: "/" + packagePath})
	}
	return packages, nil
}

func goDoc(packagePath string) (string, error) {
	command := exec.Command("go", "doc", "-all", packagePath)
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("ドキュメントの生成に失敗しました: %w\n%s", err, output.String())
	}
	return output.String(), nil
}

func render(w http.ResponseWriter, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pageTemplate.Execute(w, data); err != nil {
		log.Printf("HTMLの生成に失敗しました: %v", err)
	}
}
