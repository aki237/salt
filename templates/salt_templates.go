package templates

import (
	"github.com/aki237/salt"
	"html/template"
)

func PushTemplate(filename string, w *salt.ResponseBuffer ,fillers interface{}) (error) {
	t,err := template.ParseFiles(filename)
	if err != nil {
		return err
	}
	return t.Execute(*w, fillers)
}
