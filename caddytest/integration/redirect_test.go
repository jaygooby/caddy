package integration

import (
	"os"
	"testing"
  "fmt"
	"path/filepath"
	"github.com/caddyserver/caddy/v2/caddytest"
)

// This is the same TestRedirect as seen in
// caddytest/integration/caddyfile_test.go
func TestRedirect(t *testing.T) {

	// arrange
	tester := caddytest.NewTester(t)
	tester.InitServer(`
  {
    admin localhost:2999
    http_port     9080
    https_port    9443
    grace_period  1ns
  }

  localhost:9080 {

    redir / http://localhost:9080/hello 301

    respond /hello 200 {
      body "hello from localhost"
    }
    }
  `, "caddyfile")

	// act and assert
	tester.AssertRedirect("http://localhost:9080/", "http://localhost:9080/hello", 301)

	// follow redirect
	tester.AssertGetResponse("http://localhost:9080/", 200, "hello from localhost")
}

// This tests Test case 1 from
// https://github.com/caddyserver/caddy/issues/4205#issuecomment-863352037
func TestArticleIndexRedirect(t *testing.T) {

	// temporary dir to hold the caddy config
	// tmpdir := t.TempDir()
	tmpdir := "/tmp/caddy"

	err := os.MkdirAll(filepath.Join(tmpdir,"expert-caddy"), os.ModePerm)
	if err != nil {
		t.Errorf("can't create dir: %s", err)
  }

	// put a dummy template in there
	templatefile := []byte("template body\n")
	err = os.WriteFile(filepath.Join(tmpdir,"expert-caddy","/template.html"), templatefile, 0644)
	if err != nil {
		t.Errorf("can't write template file: %s", err)
  }

	// arrange
	tester := caddytest.NewTester(t)
	tester.InitServer(
		fmt.Sprintf(`
	  {
	    admin localhost:2999
	    http_port     9080
	    https_port    9443
	    grace_period  1ns
	  }

	  localhost:9080 {

			 root * %s

			 templates

			 # article index should have trailing slash to preserve hrefs
			 redir /expert-caddy /expert-caddy/
			 try_files {path} {path}/ {path}.html

			 # serve all articles from the template
			 rewrite /expert-caddy/* /expert-caddy/template.html

			 file_server

	  }
	  `, tmpdir), "caddyfile")

	// act and assert
	tester.AssertRedirect("http://localhost:9080/expert-caddy", "http://localhost:9080/expert-caddy/", 302)

	// can serve /expert-caddy/template.html
	tester.AssertGetResponse("http://localhost:9080/expert-caddy/template.html", 200, "template body\n")

	// all articles should be served via /expert-caddy/template.html
	tester.AssertGetResponse("http://localhost:9080/expert-caddy/foo", 200, "template body\n")
}
