package parser

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloadAndExtract(t *testing.T) {
	outDir := "."
	url := "https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html"
	resp, err := http.Get(url)
	assert.NoError(t, err)
	defer resp.Body.Close()
	err = Download(url, outDir, resp.Body)
	assert.NoError(t, err)
	_, err = Extract(resp)
	assert.NoError(t, err)
	assert.NoError(t, os.Remove(filepath.Join(outDir, format(url))))
}

func TestFormat(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{
			in:  "http://www.baidu.com",
			out: "www.baidu.com",
		},
		{
			in:  "https://cloud.baidu.com/BFE",
			out: "cloud.baidu.com_BFE",
		},
		{
			in:  "http://family.com/help_index.html",
			out: "family.com_help-index.html",
		},
	}
	for i := range cases {
		assert.Equal(t, cases[i].out, format(cases[i].in))
	}
}
