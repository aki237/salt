package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	RED = "\x1b[31;1m"
	GREEN = "\x1b[32;1m"
	YELLOW = "\x1b[33;1m"
	BLUE = "\x1b[34;1m"
	PINK = "\x1b[35;1m"
	AQUA = "\x1b[36;1m"
	NORMAL = "\x1b[0m"
)

var gitchanged string

var errormsg string = `Salt : the salt web app manager
Usage : salt <command> [<command args>...]

List of commands :
    create <webappname> [<webappdir>] - Create a new web app skeleton
    run <webappname> - Run a created web app`

var app_json string =
`{
    "Debug" : true,
    "ListenVars" : {
        "Port" : "8080",
        "Address": "127.0.0.1"
    },
    "Static" : {
        "StaticURI" : "/static/",
        "StaticDirs" : ["static/"]
    },
    "Database" : {
        "Username" : ${Username},
        "Password" : ${Password},
        "Database" : ${Database}
    }
}`

var appname_go string =
`package main

import (
	"./{{appname}}"
	"github.com/aki237/salt"
)

func main(){
	salt.Configure("app.json")
	salt.Add404(NotFound)
	salt.AddRootApp({{appname}}.App)
	salt.Run()
}

//Not found function
func NotFound(w salt.ResponseBuffer, r *salt.RequestBuffer)  {
	w.Write([]byte("The page you are looking for doesn't exist"))
}`

var appname_urls_go string =`package {{appname}}

import "github.com/aki237/salt"

var URLS salt.URLS = salt.URLS{
	{Routename : "home",Pattern : "/$",Handler : salt.SampleHome ,},
}`

var appname_app_go string =`package {{appname}}

import "github.com/aki237/salt"

var App salt.App = salt.App{
	URLS:URLS,
	Models:Models,
	BaseURL:"/",
}
`
var appname_views_go string = `package {{appname}}

import "github.com/aki237/salt"

//YOUR VIEWS GO HERE
`

var appname_models_go string = `package {{appname}}

import "github.com/aki237/salt/models"

//SAMPLE MODEL ARRAY
var Models models.Models = models.Models{
	models.Model{
		Name : "NEWMODEL",
		Fields: models.Fields{
			"ID":models.Field{models.Integer,true,true},
			"MODEL_ENTITY_1":models.Field{models.CharField,false,false},
			"MODEL_ENTITY_2":models.Field{models.CharField,false,false},
			"MODEL_ENTITY_3":models.Field{models.CharField,false,false},
		},
		PrimaryKey : "ID",
	},
}
`

var wd string
func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println(errormsg)
		return
	}
	switch args[1]{
	case "create":
		if (len(args) < 3) {
			fmt.Println(RED + "Wrong Usage of create command" + NORMAL + "\n" + errormsg)
			return
		}
		create(args[2:])
		return
	case "run":
		if (len(args) < 3) {
			fmt.Println(RED + "Wrong Usage of run command" + NORMAL + "\n" + errormsg)
			return
		}
		run(args[2:])
		return
	default:
		fmt.Println(RED + args[1] + NORMAL + " : Command not found")
		return
	}
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

//
func create(args []string)  {
	wd,_ = os.Getwd()
	if (len(args) == 2){
		wd = args[1]
	}
	
	webappname := args[0]
	fmt.Println(YELLOW+"Salt : Shaking . . . "+NORMAL)
	fmt.Println(BLUE + "Web Appname : "+ GREEN + webappname +NORMAL)
	if isexist , _ := exists(wd) ; !isexist {
		fmt.Println(RED + wd + " : Directory doesn't exist")
		return
	}
	fmt.Println(BLUE + "Web App Directory : "+ GREEN + wd + NORMAL)
	if isexist, _ := exists(wd+"/"+webappname); isexist {
		fmt.Println(RED + "A file/dir with the same name already exists")
	}
	err := os.MkdirAll(wd + "/" + webappname + "/" + webappname, 0755)
	if err != nil {
		fmt.Println(RED,err)
		return
	}
	err = os.MkdirAll(wd + "/" + webappname + "/static", 0755)
	if err != nil {
		fmt.Println(RED,err)
		return
	}
	app_json = Replace(app_json, webappname)
	err = ioutil.WriteFile(wd + "/" + webappname + "/app.json",[]byte(app_json), 0644)
	if err != nil {
		fmt.Println(RED,err)
		return
	}
	appname_go = Replace(appname_go, webappname)
	err = ioutil.WriteFile(wd + "/" + webappname + "/" + webappname + ".go",[]byte(appname_go), 0644)
	if err != nil {
		fmt.Println(RED,err)
		return
	}
	appname_urls_go = Replace(appname_urls_go, webappname)
	err = ioutil.WriteFile(wd + "/" + webappname + "/" + webappname + "/urls.go",[]byte(appname_urls_go), 0644)
	if err != nil {
		fmt.Println(RED,err)
		return
	}
	appname_app_go = Replace(appname_app_go, webappname)
	err = ioutil.WriteFile(wd + "/" + webappname + "/" + webappname + "/app.go",[]byte(appname_app_go), 0644)
	if err != nil {
		fmt.Println(RED,err)
		return
	}
	appname_views_go = Replace(appname_views_go, webappname)
	err = ioutil.WriteFile(wd + "/" + webappname + "/" + webappname + "/views.go",[]byte(appname_views_go), 0644)
	if err != nil {
		fmt.Println(RED,err)
		return
	}
	appname_models_go = Replace(appname_models_go, webappname)
	err = ioutil.WriteFile(wd + "/" + webappname + "/" + webappname + "/models.go",[]byte(appname_models_go), 0644)
	if err != nil {
		fmt.Println(RED,err)
		return
	}
	fmt.Println(BLUE + "All done... now goto that dir and run : "+GREEN+"salt run "+webappname+".go"+NORMAL)
}

//
func run(args []string){
	wd ,_ = os.Getwd()
	appname := args[0]
	if isexist , _ := exists(wd+"/"+appname+".go") ; !isexist {
		fmt.Println(RED + appname +" : Unable to find the app files.")
		return
	}
	cmd := exec.Command("go", "run",appname+".go")
	stdout , err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s%s\t %s>>%s %s%s\n",YELLOW,time.Now().String(),GREEN,BLUE, scanner.Text(),NORMAL)
		}
	}()
	cmd.Run()
}

//
func Replace(a string ,appname string) (string) {
	return strings.Replace(a,"{{appname}}",appname,-1)
}
