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
	tmpdir := t.TempDir()

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

// Test case 2 from
// https://github.com/caddyserver/caddy/issues/4205#issuecomment-863352037
func TestNonCanonicalizedRewrittenFilename(t *testing.T) {

	// temporary dir to hold the caddy config
	tmpdir := t.TempDir()

	// make the various nested dirs and their index.html
	dirs := []string{"json", "modules"}
	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(tmpdir,"docs",dir), os.ModePerm)
		if err != nil {
			t.Errorf("can't create dir: %s", err)
	  }
		// content is the same as the name of the file for testing purposes
		err = os.WriteFile(filepath.Join(tmpdir,"docs",dir, "index.html"), []byte(dir), 0644)
		if err != nil {
			t.Errorf("can't write index file: %s", err)
	  }
	}

	// make the /docs/index.html file
	err := os.WriteFile(filepath.Join(tmpdir,"docs","index.html"), []byte("docs"), 0644)
	if err != nil {
		t.Errorf("can't write index file: %s", err)
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

			 file_server
       templates
       encode gzip

       try_files {path}.html {path}

       redir   /docs/json      /docs/json/
       redir   /docs/modules   /docs/modules/
       rewrite /docs/json/*    /docs/json/index.html
       rewrite /docs/modules/* /docs/modules/index.html
       rewrite /docs/*         /docs/index.html

       reverse_proxy /api/* localhost:4444
	  }
	  `, tmpdir), "caddyfile")

	// redir /docs/json
	tester.AssertRedirect("http://localhost:9080/docs/json", "http://localhost:9080/docs/json/", 302)

	// redir /docs/modules
	tester.AssertRedirect("http://localhost:9080/docs/modules", "http://localhost:9080/docs/modules/", 302)

	// index.html in /docs/json/ serves any file
	tester.AssertGetResponse("http://localhost:9080/docs/json/foo", 200, "json")

	// index.html in /docs/modules/ serves any file
	tester.AssertGetResponse("http://localhost:9080/docs/modules/wat", 200, "modules")

	// index.html in /docs/ serves any file except for json and modules
	tester.AssertGetResponse("http://localhost:9080/docs/weevils", 200, "docs")

	// index.html in /docs/ serves any file except for json and modules including
	// nested paths
	tester.AssertGetResponse("http://localhost:9080/docs/weevils/grub", 200, "docs")
}
