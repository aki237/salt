# salt : yet another tasty {M}VC for golang

## Intro
Salt is a go package to improve routing upon the standard http package. It has inbuilt mechanisms to maintain models and connect to databases(only MySQL for now).

## Quickstart
This repo contains a salt cmd tool.
```shell
salt create [appname]
```
It creates a folder of the same name with a sample app contained in it.
The file app.json is the app configuration for the web application.
```json
{
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
        "Username" : "Username",
        "Password" : "Password",
        "Database" : "Database"
    }
}
```
The json fields are self explainatory.

To run it either
```shell
salt run [appname]
```
or
```shell
go run [appname].go
```
inside that folder.

Documentation is under process. For now refer to the godoc page. For further issues, well, use the github issue utility to file any.
### For contributing : See the [Tasks](TASKS.md)
