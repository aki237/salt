//Package that implements the MVC framework in golang.
//The salt/salt package which is an executable can be a good kick-starter for scaffolding a small salt web-app which
//can show a sample home page.
package salt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"strings"
	"github.com/aki237/salt/models"
)

//The struct containing the information about all the url patterns
type URL struct {
	Pattern string
	Routename string
	Handler func(ResponseBuffer, *RequestBuffer)
}

//Array of the URL Type
type URLS []URL

//App structure
type App struct {
	URLS    URLS
	Models  models.Models
	BaseURL string
}

//Struct for storing the config read in the app.json file of the app
type Config struct{
	Debug bool 
	ListenVars struct{
		Port string 
		Address string 
	}
	Static struct{
		StaticURI string 
		StaticDirs []string 
	}
	Database models.Database
}

//The runtime variable : config - containing the configuration of an web-app read from the file passed
//as a parameter to the Configure function
var config Config

//Boolean variable that states whether the web-app is configured with a Configure function call successfully
var configured bool = false

//Boolean variable that states whether a root app has been added web-app
var rootapppresent bool = false


//The function that actually listens to the listenVar (address:port or : port).
//Here the server's root url is mapped to a default router function salt.router
//This router inturn matches the urls with the registered patterns and Runs the Handler Function (in the Route struct).
//After registering the routes, execute this functon to start the server.
func Run() error {
	http.HandleFunc("/", router)
	return http.ListenAndServe(config.ListenVars.Address+":"+config.ListenVars.Port, nil)
}

func RunAt(serveaddr string) error {
	http.HandleFunc("/", router)
	return http.ListenAndServe(serveaddr, nil)
}

//This is the function that serves the static files present in the directories specified in the
//configuration file at the static URL pattern (which again should also be specified in the config file).
func StaticServe(w ResponseBuffer, r *RequestBuffer)  {
	filename := r.URLParameters["staticfile"].(string)
	log(filename)
	filename = strings.Replace(filename , "../","_",-1)
	filename = strings.Replace(filename , "./","/",-1)
	extention := strings.Split(filename, ".")[len(strings.Split(filename, ".")) - 1]
	
	for _,entry := range config.Static.StaticDirs {
		if exists(entry + "/" + filename) {
			contenttype := mime.TypeByExtension("."+extention)
			w.Header().Set("Content-Type", contenttype)
			filebyte,err := ioutil.ReadFile(entry + "/" + filename)
			if err != nil{
				fmt.Fprint(w,err.Error())
			} else {
				log(filename," found in ",entry)
				w.Write(filebyte)
			}
			return
		}
	}
	
	Func404(w,r)	
}

//This function is used to configure a particular web-app
func Configure(filename string)(error)  {
	
	content, err := ioutil.ReadFile(filename)
	if(err != nil){
		return err
	}

	
	err = json.Unmarshal(content,&config)
	if (len(config.Static.StaticURI) > 0){
		log("Static File Directories detected : " , config.Static.StaticDirs)
		if (string(config.Static.StaticURI[0]) != "/"){
			config.Static.StaticURI = "^/" + config.Static.StaticURI
		} else {
			config.Static.StaticURI = "^" + config.Static.StaticURI			
		}

		
		if (string(config.Static.StaticURI[len(config.Static.StaticURI)-1]) != "/"){
			config.Static.StaticURI = config.Static.StaticURI + "/"
		}
		var staticdirs []string
		for _, dirs := range config.Static.StaticDirs {
			stat,err := os.Stat(dirs)
			if (os.IsNotExist(err)){
				log("The specified static directory doesn't exist : " + dirs)
			} else {
				if (!stat.IsDir()){
					log("The specified static entry is not a directory : " + dirs)
				} else {
					staticdirs = append(staticdirs,dirs)
				}
			}
		}
		config.Static.StaticDirs = staticdirs
		static := URL{
			Pattern : config.Static.StaticURI + "<any:staticfile>",
			Routename : "Static",
			Handler : StaticServe,
		}
		static.AddRoute()
		
	}
	fmt.Println(config)
	err = models.SetDatabaseConfig(config.Database)
	if (err == nil){
		configured = true
	}
	return err
}

//Add routes from an array of URL
func (routes URLS)AddRoutes()  {
	for _, val := range routes{
		val.AddRoute()
	}
}



//Add a route from a URL variable
func (routeconf URL)AddRoute()  {
	err := AddRoute(routeconf.Pattern, routeconf.Routename, routeconf.Handler)
	if err != nil {
		log(err)
	}
}



//This function is used to add a route app to the web-app. All salt web-app should have one (and only) root app.
//An app here refers to the collection of urls, views and models.
func AddRootApp(app App) (error) {
	if !(rootapppresent) && (configured){
		app.URLS.AddRoutes()
		rootapppresent = true
		fmt.Println("Registering App Models ...")
		for _, val := range app.Models {
			fmt.Println(val.Name)
			if !val.IsMigrated() {
				fmt.Println(val.Name," is not migrated yet.Migrating...")
				return val.Register()
			} else {
				fmt.Println("Already Migrated")
			}
		}
		return nil
	}
	if (!configured){
		return errors.New("The app is not configured")
	}
	return errors.New("A root app has already been registered. Try the URLS with AddApp.")
}




//This function is used to add a new app to the web-app. A web-app can contian more than one auxillary apps. 
func AddApp(app App) (error)  {
	if (!configured){
		return errors.New("The app is not configured")
	}
	if (!rootapppresent) {
		return errors.New("Root app has not been registered.")
	}
	for _, val := range routes {

		if (len(app.BaseURL) < len(val.Pattern)){

			if (app.BaseURL == val.Pattern[:len(app.BaseURL)]){

				log("The /api/ base Pattern has already been used in the root app. This new app may not work as expected.")
				break

			}
		}
	}

	for index,_ := range app.URLS {
		app.URLS[index].Pattern = app.BaseURL + app.URLS[index].Pattern
	}
	
	app.URLS.AddRoutes()
	return app.Models.Register()
}



//This is used to get whether a file exists in a path.
func exists(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

//Sample Home Page
//This is the function is used to display the Sample Page that appears after making the new app from
//the salt commandline tool.
func SampleHome(w ResponseBuffer, r *RequestBuffer)  {
	homecontent := `	<html>
	<head>
		<meta charset="UTF-8"></meta>
		<title>salt - It's working</title>
	</head>
	<style type="text/css">
	body{
		background-color: #27282;
	}
	#container{
		position: absolute;
		top : 0px;
		background-color: #FFF;
		left: 0px;

		width: 100%;
	}
	#header{
		font-weight:100;
		text-align: center;
		top: 0px;
		left: 0px;
		border-top:#aeaeae dashed 5px;
		border-left:#aeaeae dashed 5px;
		border-right:#aeaeae dashed 5px;
		font-size: 40px;
		background-color: #E0EBFF;
		color: #3e3e3e;
	}
	#content{
		font-weight:100;
		background-color: #EEEEEE;
		border-width: 5px;
		font-size: 30px;
		color: #aeaeae;
		height: 500px;
		padding: 20px;
		border-style: dashed;
	}
	a{
		text-decoration: none;
		color: #aeaeae;
		font-weight: bold;
	}
	a:hover{
		color: #000000;
		border-bottom: #aeaeae dashed 1px;
	}
	a:active{
		color: #aaaaaa;
		border-bottom: #aeaeae dashed 1px;
	}
	a:visited{
		color:#bbbbbb;
		background-color: #aeaeae;
	}
	</style>
	<body>
		<div id="container">
			<div id="header">salt : It's working</div>
			<div id="content">
				Now add new routes and write corresponding functions.
				For more refer to the documention of <a href="https://github.com/aki237/salt">salt</a>.
				This is the default page for all new salt webapps or You haven't configured your app properly.
                                This page may also come up if no routes are registered in the app.
			</div>
		</div>
	</body>
	</html>`
	fmt.Fprint(w, homecontent)
	
}


//This is the function is used to log the output of salt app internals. The Debug can be turned false in the web-app
//configuration file.
func log(a ...interface{})  {
	if (config.Debug){
		fmt.Println(a)
	}
}
