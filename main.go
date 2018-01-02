package main

import (
	"flag"
	"github.com/kataras/iris"
	"fmt"
	"os"
	"github.com/inconshreveable/log15"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"encoding/json"
)

const (
	configFilename = "config.yaml"
)

var conf *Config

func main() {

	servePort := flag.Int("bind", 0, "Http bind port")
	devMode := flag.Bool("dev", false, "Is dev mode?")
	flag.Parse()

	if *servePort != 0 {
		serve(*servePort, *devMode)
	}
}

func serve(bind int, devMode bool) {
	s := iris.New()
	s.RegisterView(iris.HTML("templates", ".html").Reload(devMode))
	app := s.Party("/api/v1")
	files, err := ioutil.ReadDir("./templates")
	if err != nil {
		log15.Error(err.Error())
		log15.Error("fail to open templates directory")
		os.Exit(2)
	}
	conf = getConf()
	for _, file := range files {
		// avoid range change value affect callback value
		f := file
		if f.IsDir() && f.Name()[:1] != "." {
			app.StaticWeb(fmt.Sprintf("/template/%s/images", f.Name()), fmt.Sprintf("./templates/%s/images", f.Name()))
			app.Get(fmt.Sprintf("/template/%s", f.Name()), func(ctx iris.Context) {
				getTemplate(ctx, f.Name(), devMode)
			})
			app.Post(fmt.Sprintf("/template/%s", f.Name()), func(ctx iris.Context) {
				getTemplateFromData(ctx, f.Name(), devMode)
			})
			log15.Info(fmt.Sprintf("add directory %s", f.Name()))
		}
	}
	s.Run(iris.Addr(fmt.Sprintf(":%d", bind)), iris.WithoutVersionChecker)
}

type Config struct {
	Vars         map[string]string `yaml:"vars"`
	Social       map[string]string `yaml:"social"`
	Link         map[string]string `yaml:"link"`
	ProdRootPath string            `yaml:"prod_root_path"`
}

func getConf() *Config {
	data, err := ioutil.ReadFile(fmt.Sprintf("./%s", configFilename))
	if err != nil {
		log15.Error(err.Error())
		log15.Error("fail to open config file")
		os.Exit(2)
	}
	var config = new(Config)
	err = yaml.Unmarshal(data, config)
	if err != nil {
		log15.Error(err.Error())
		log15.Error("Fail to open config file")
		os.Exit(2)
	}
	return config
}

func getTemplate(ctx iris.Context, directory string, devMode bool) {
	var data map[string]interface{}
	rawData, err := ioutil.ReadFile(fmt.Sprintf("./templates/%s/data.json", directory))
	if err == nil {
		err := json.Unmarshal(rawData, &data)
		if err != nil {
			log15.Error(err.Error())
		}
	} else {
		log15.Error(err.Error())
	}
	ctx.ViewData("data", data)

	ctx.ContentType("text/html")
	ctx.Header("cache-control", "no-cache")
	ctx.Header("pragma", "no-cache")
	ctx.Header("expires", "-1")
	ctx.ViewData("conf", conf)
	rootPath := conf.ProdRootPath
	if devMode {
		rootPath = directory
	}
	ctx.ViewData("rootPath", rootPath)
	ctx.View(fmt.Sprintf("%s/index.html", directory))
}

func getTemplateFromData(ctx iris.Context, directory string, devMode bool) {
	var data interface{}
	err := ctx.ReadJSON(&data)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}
	ctx.ContentType("text/html")
	ctx.Header("cache-control", "no-cache")
	ctx.Header("pragma", "no-cache")
	ctx.Header("expires", "-1")
	ctx.ViewData("conf", conf)
	ctx.ViewData("data", data)
	rootPath := conf.ProdRootPath
	if devMode {
		rootPath = directory
	}
	ctx.ViewData("rootPath", rootPath)
	ctx.View(fmt.Sprintf("%s/index.html", directory))
}
