package core

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/TOsmanov/go-pdf/internal/config"
	"github.com/stretchr/testify/assert"
)

const (
	logoPath              = "../test_assets/logo.jpg"
	testTitle             = "The Go Programming Language"
	footerTitle           = "Sub-title text. 2024 year"
	templateTitlePagePath = "../templates/title-page.html"
	templateTocPagePath   = "../templates/toc-page.html"
	templateFacePagePath  = "../face/index.html"
)

func TestToCPageHTML(t *testing.T) {
	var b string
	var cfg config.Config
	var page Page

	params := Params{
		Title:          "Main page of the documentation. Very important documentation, very important documentation",
		FooterTitle:    footerTitle,
		SubTitle:       "Specification",
		Landscape:      "0",
		FinalLandscape: false,
	}
	cfg.LogoPath = logoPath
	cfg.Templates.TitlePage = templateTitlePagePath
	cfg.Templates.TocPage = templateTocPagePath
	cfg.Face.FacePage = templateFacePagePath

	err := cfg.InitTemplates()
	assert.Nil(t, err)

	pageToC := `{"title":"Golang Docs","anchor":"gopdf-page-header-00001",
	"headers":[{"lvl":2,"text":"Download and Install", "link": "download"},
	{"lvl":3,"text":"Binary Distributions","link":"link1"},
	{"lvl":3,"text":"Install From Source","link":"link2"},
	{"lvl":2,"text":"Contributing","link":"link3"},
	{"lvl":3,"text":"First header","link":"link4"},
	{"lvl":3,"text":"Second header","link":"link5"}]}`
	err = json.Unmarshal([]byte(pageToC), &page)
	assert.Nil(t, err)
	params.ToC = append(params.ToC, page)

	err = ToCPageHTML(&params, &b, &cfg)
	assert.Nil(t, err)

	expected := `<style>
div.toc-pages {
    color: black;
    a {
        color: black;
        text-decoration: none;
    }
    ul {
        display: block;
        list-style-type: none;
        page-break-inside: avoid;
        margin-left: 0;
        padding-left: 0;
        li {
            list-style: none;
            display: block;
        }
    }
    div.toc-page, h3 {
        page-break-after: avoid;
        page-break-inside: avoid;
        break-after: avoid;
    }
}
</style>
<div class="toc-pages">
    <div class="toc-page" style="margin-bottom:50px;">
        <span class="h3" style="margin-bottom:20px; page-break-after:avoid;font-weight: 700;">1   <a href="#gopdf-page-header-00001">Golang Docs</a></span>
        <ul style="page-break-before:avoid; margin-top:8px;">
            <li style="padding-left: calc(-1ch);"><a href="#download">1.1    Download and Install</a></li>
            <li style="padding-left: calc(4ch);"><a href="#link1">1.1.1    Binary Distributions</a></li>
            <li style="padding-left: calc(4ch);"><a href="#link2">1.1.2    Install From Source</a></li>
            <li style="padding-left: calc(-1ch);"><a href="#link3">1.2    Contributing</a></li>
            <li style="padding-left: calc(4ch);"><a href="#link4">1.2.1    First header</a></li>
            <li style="padding-left: calc(4ch);"><a href="#link5">1.2.2    Second header</a></li>
            </ul>
    </div>
    <div style="page-break-after: always;"></div>
</div>
`
	// fmt.Println(b)
	assert.Equal(t, expected, b)
}

func TestPrepareToC(t *testing.T) {
	var params Params
	var page Page

	pageToC := `{"title":"Golang Docs","anchor":"gopdf-page-header-00001",
	"headers":[
	{"lvl":2,"text":"Download and Install"},
	{"lvl":3,"text":"Binary Distributions"},
	{"lvl":3,"text":"Install From Source"},
	{"lvl":4,"text":"Test 4 lvl"},
	{"lvl":4,"text":"Test 4 lvl"},
	{"lvl":5,"text":"Test 5 lvl"},
	{"lvl":2,"text":"Download and Install"},
	{"lvl":3,"text":"Binary Distributions"},
	{"lvl":3,"text":"Install From Source"},
	{"lvl":4,"text":"Test 4 lvl"},
	{"lvl":4,"text":"Test 4 lvl"},
	{"lvl":5,"text":"Test 5 lvl"},
	{"lvl":2,"text":"Contributing"},
	{"lvl":3,"text":"First header"},
	{"lvl":3,"text":"Second header"}]}`
	err := json.Unmarshal([]byte(pageToC), &page)
	assert.Nil(t, err)
	params.ToC = append(params.ToC, page)

	PrepareToC(&params)
	b, err := json.Marshal(params.ToC)
	assert.Nil(t, err)
	excepted := `[{"index":1,"title":"Golang Docs","anchor":"gopdf-page-header-00001",
	"headers":[{"lvl":2,"text":"Download and Install","link":"","index":"1.1","class":"lvl-2","indent":-1},
	{"lvl":3,"text":"Binary Distributions","link":"","index":"1.1.1","class":"lvl-3","indent":4},
	{"lvl":3,"text":"Install From Source","link":"","index":"1.1.2","class":"lvl-3","indent":4},
	{"lvl":4,"text":"Test 4 lvl","link":"","index":"1.1.2.1","class":"lvl-4","indent":10},
	{"lvl":4,"text":"Test 4 lvl","link":"","index":"1.1.2.2","class":"lvl-4","indent":10},
	{"lvl":5,"text":"Test 5 lvl","link":"","index":"1.1.2.2.1","class":"lvl-5","indent":17},
	{"lvl":2,"text":"Download and Install","link":"","index":"1.2","class":"lvl-2","indent":-1},
	{"lvl":3,"text":"Binary Distributions","link":"","index":"1.2.1","class":"lvl-3","indent":4},
	{"lvl":3,"text":"Install From Source","link":"","index":"1.2.2","class":"lvl-3","indent":4},
	{"lvl":4,"text":"Test 4 lvl","link":"","index":"1.2.2.1","class":"lvl-4","indent":10},
	{"lvl":4,"text":"Test 4 lvl","link":"","index":"1.2.2.2","class":"lvl-4","indent":10},
	{"lvl":5,"text":"Test 5 lvl","link":"","index":"1.2.2.2.1","class":"lvl-5","indent":17},
	{"lvl":2,"text":"Contributing","link":"","index":"1.3","class":"lvl-2","indent":-1},
	{"lvl":3,"text":"First header","link":"","index":"1.3.1","class":"lvl-3","indent":4},
	{"lvl":3,"text":"Second header","link":"","index":"1.3.2","class":"lvl-3","indent":4}]}]`
	excepted = regexp.MustCompile(`,\n\t*`).ReplaceAllString(excepted, ",")
	// fmt.Println(string(b))
	assert.Equal(t, excepted, string(b))
}

func TestTitlePageHTML(t *testing.T) {
	var b string
	var cfg config.Config

	params := Params{
		Title:          "Main page of the documentation. Very important documentation, very important documentation",
		FooterTitle:    footerTitle,
		SubTitle:       "Specification",
		Landscape:      "0",
		FinalLandscape: false,
	}
	cfg.LogoPath = logoPath
	cfg.Templates.TitlePage = templateTitlePagePath
	cfg.Templates.TocPage = templateTocPagePath
	cfg.Face.FacePage = templateFacePagePath

	err := cfg.InitTemplates()
	assert.Nil(t, err)

	err = TitlePageHTML(&params, &b, &cfg)
	assert.Nil(t, err)

	expected := `<div class="title-page" style="display: block; height: 800px; margin-left: 70px; margin-bottom: 100px;">
    <h1 style="font-weight: 700; margin-top: 50px">Main page of the documentation. Very important documentation, very important documentation</h1>
    <h3 style="margin-top: 100px;">Specification</h3>
        </div>
<div style="display: block; margin-left: 70px;"><h5>Sub-title text. 2024 year</h5></div>
<div style="page-break-after: always;"></div>
`
	// fmt.Println(b)
	assert.Equal(t, expected, b)
}
