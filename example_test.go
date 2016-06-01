package salt_test

import (
	"github.com/aki237/salt"
	"fmt"
)

func ExampleAddRoute(){
	salt.AddRoute("/<all:user>/<str:post>/<int:commentid>$","commentid",showcomment)
	salt.AddRoute("/<all:user>/<str:post>$","post",showpost)
	// Output:
	// pattern =  /(?P<user>[[:alnum:]]+)/(?P<post>[[:alpha:]]+)/(?P<commentid>[[:digit:]]+)$
	// pattern =  /(?P<user>[[:alnum:]]+)/(?P<post>[[:alpha:]]+)
}

func ExampleModifyRoute(){
	//Here only the name of the route and the patterns are changed and not the handler function, which can be done too.
	salt.ModifyRoute("post", "showpost", "/<all:user>/<str:post>/show$", showpost)
	// Output:
	// pattern =  /(?P<user>[[:alnum:]]+)/(?P<post>[[:alpha:]]+)/show
}

//
func ExampleValidate() {
	exp,err := salt.Validate("/api/<int:userid>")
	if err != nil {
		//Do the error handling.
		fmt.Println(err)
	}
	// Output:
	// pattern =  /api/(?P<userid>[[:alnum:]]+)
}

func ExampleAddNewRouteObject(){
	exp,err := salt.Validate("/api/<int:userid>")
	if err != nil {
		//Do the error handling.
		fmt.Println(err)
	}
	newroute := salt.Route{exp,"/api/<int:userid>",handler,"userdetails"}
	if newroute.AddNewRouteObject() != nil {
		//Do the error handling.
		fmt.Println(err)
	}
	// Output:
	// If the route is valid :
	// <nil>
	// <nil>
	// Else :
	// Error : .....
	// Error : .....
}
