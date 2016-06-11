//Package that implements the regexp URL handling upon the http standard package
//
//The standard http package implements most of the server based functions. But the default route handler
// ie., the http.HandleFunc is unable to handle urls with regexp as in Django or any other web framework.
//This package implements the regexp routing of urls upon the standard net/http concepts.
//The following is a simple program that shows how this package operates.
/*
   package main

   import (
           "fmt"
           "github.com/aki237/salt"
   )

   func main() {
       salt.AddRoute("/<all:username>/posts/<int:postno>$","post",getpost)
       salt.AddRoute("/<all:username>$","user",getuser)
       salt.Run(":1080")
   }

   func getpost(w salt.ResponseBuffer, r *salt.RequestBuffer){
       username := r.URLParameters["username"].(string)
       postno := r.URLParameters["postno"].(int)
       fmt.Fprintln(w,username,postno)
   }

   func getuser(w salt.ResponseBuffer, r *salt.RequestBuffer){
       username := r.URLParameters["username"].(string)
       fmt.Fprintln(w,username)
   }
*/
// This program runs a simple server at localhost:1080.
//The URL http://127.0.0.1:1080/aki237, will show "aki237".
//The URL http://127.0.0.1:1080/aki237/post/898, will show "aki237 898".
package salt



import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)



//RegexpMap is the type that inherits the regexp.Regexp Struct and contains the variable type of different url pattern variables
type RegexpMap struct {
	*regexp.Regexp
	typeMaps map[string]string
}



//Route is the type that contains routing information for different urls registered
type Route struct {
	RegexpPattern *RegexpMap
	Pattern       string
	Handler       func(ResponseBuffer, *RequestBuffer)
	Name          string
}



//This type is directly inherited from the http.ResponseWriter. For more refer http.ResponseWriter from the standard package.
type ResponseBuffer http.ResponseWriter



//This type is an extension of the http.Request package. In addition to the http.Request element, it also contains a map variable
// of an interface mapped to names(string) of the URLPattern variables.
type RequestBuffer struct {
	*http.Request
	error         error
	URLParameters map[string]interface{}
}



//Global Private variable that contains all the registered routes
var routes []Route

var Func404 func(w ResponseBuffer , r *RequestBuffer) = Default404

//
func Default404(w ResponseBuffer , r *RequestBuffer)  {
	fmt.Fprint(w,"Page not found")
}

//
func Add404(Handler  func(ResponseBuffer, *RequestBuffer))  {
	Func404 = Handler
}

//This router function is the default router of root url of the server. Other URLs are routed from here.
func router(w http.ResponseWriter, r *http.Request) {
	urlstr := r.URL.EscapedPath()
	tmp := make(map[string]interface{}, 1)
	var err error
	if (len(routes) == 0){
		SampleHome(w,&RequestBuffer{r,nil,nil})
	}
	for _, route := range routes {
		if route.RegexpPattern.MatchString(urlstr) {
			temp := &RequestBuffer{r, err, tmp}
			for _, mapname := range route.RegexpPattern.SubexpNames()[1:] {
				switch route.RegexpPattern.typeMaps[mapname] {
				case "str", "all","any":
					temp.URLParameters[mapname] = route.RegexpPattern.ReplaceAllString(urlstr, "${"+mapname+"}")
				case "int":
					temp.URLParameters[mapname], temp.error = strconv.Atoi(route.RegexpPattern.ReplaceAllString(urlstr, "${"+mapname+"}"))
				}
			}
			route.Handler(w, temp)
			return
		}
	}
	Func404(w, &RequestBuffer{r,nil,nil})
}




//This function is used to add new routes to the server. As the function defenition suggests the function should be called
// with the given parameters :
//
// +  pattern  :  The url pattern
//    - All the regexp patterns are appliable ($ - end, ^ - begining and etc.,)
//    - Constructing a pattern is made simple : <[type]:[varable_name]>
//    - Allowed types are :
//           * str - Only alphabet class = [[:alpha:]]
//           * int - Only the number Class = [[:digit:]]
//           * all - Class formed by str + int = [[:alnum:]]
//    - Variables are only constructed using Alphabets.
/*    Example
 *  /<all:username>$                  translates to regexp pattern /(?P<username>[[:alnum:]]+)$
 *  /<all:username>/<str:reponame>    translates to regexp pattern /(?P<username>[[:alnum:]]+)/(?P<reponame>[[:alpha:]]+)*/
//    - URL Parameter variables cannot be used more than once in a url :
//           *  /<all:username>/<str:username>    is invalid
// +  routename  :  The name of the route
//    - This a simple string that identifies a route.
//    - This is unique for a route. So registering another route with a same name is invalid. That route will not be added returnin an error.
//    - This name is used to modify any route during the runtime using the ModifyRoute function in this package.
// +  handler  -  The function which has to be called when the url pattern matches with the registered routes, with the request and the response buffers as parameters.
//    - This is similar to the handler passed to http.HandleFunc but with the modified structs ResponseBuffer and RequestBuffer.
func AddRoute(pattern string, routename string, handler func(ResponseBuffer, *RequestBuffer)) (error) {
	exp, err := Validate(pattern)
	if err != nil {
		return err
	}
	for _, route := range routes {
		if routename == route.Name {
			return errors.New("The Name for this route is already used")
		}
	}
	routes = append(routes, Route{exp, pattern, handler, routename})
	return nil
}




//Validate function is used to create a valid RegexpMap struct from the Pattern passed.
func Validate(pattern string) (*RegexpMap, error) {
	types := []string{"str", "int", "all","any"}
	var regstr string
	typeMaps := make(map[string]string, 1)

	for _, kind := range types {
		restr := regexp.MustCompile("<" + kind + ":[[:alpha:]]+>")
		matches := restr.FindAllString(pattern, -1)

		for _, val := range matches {
			mapstr := strings.Replace(val, "<"+kind+":", "", -1)
			mapstr = strings.Replace(mapstr, ">", "", -1)
			for _, check := range matches {
				if strings.Count(check, mapstr) > 1 {
					return nil, errors.New("Variable used twice")
				}
			}
			typeMaps[mapstr] = kind
			switch kind {
			case "str":
				regstr = "alpha"
			case "int":
				regstr = "digit"
			case "all":
				regstr = "alnum"
			case "any":
				regstr = "."
			}
			if (regstr == "."){
				pattern = strings.Replace(pattern, "<"+kind+":"+mapstr+">", "(?P<"+mapstr+">"+regstr+"+)", -1)
			} else {
				pattern = strings.Replace(pattern, "<"+kind+":"+mapstr+">", "(?P<"+mapstr+">[[:"+regstr+":]]+)", -1)
			}
		}

	}

	re := regexp.MustCompile(pattern)
	log("pattern = ", re.String())
	newStruct := &RegexpMap{re, typeMaps}
	return newStruct, nil
}




//This function is used to modify the routes registered. The routes are identified by the name given to them earlier.
//If the route's name is not to be changed the second paraeter should be the same as the first.
//As the parameter names suggest :
//
// + oldname   -   The current name of the router.
//
// + newname   -   The new name for the route.
//
// + pattern   -   New pattern string.
//
// + handler   -   The new handler for the pattern.
func ModifyRoute(oldname string, newname string, pattern string, handler func(ResponseBuffer, *RequestBuffer)) error {
	var index int
	for index, _ = range routes {
		if routes[index].Name == oldname {
			break
		}
	}
	exp, err := Validate(pattern)
	if err != nil {
		return err
	}
	routes[index] = Route{exp, pattern, handler, newname}
	return nil
}




//This function is used to register custom Route Variable to the routes variable.
func (newroute Route) AddNewRouteObject() error {
	for index, _ := range routes {
		if routes[index].Name == newroute.Name {
			return errors.New("There already exist a route with the same name.")
		}
	}
	routes = append(routes, newroute)
	return nil
}

//GetFormValue returns the form value for the given name "key" and error.
func (r *RequestBuffer) GetFormValue(key string) (string,error) {
	err := r.ParseForm()
	if err != nil {
		return "",err
	}

	return r.FormValue(key),nil
}
