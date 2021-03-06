package web

//go:generate go run .packr/packr.go

import (
	"encoding/json"
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"text/template"
)

type GoTemplate struct {
	Schema string
	Form   string
}

func RunWebServer(config Configuration) error {
	//Redirect requests from known locations to the embedded content from ./frontend
	box := packr.New("frontend", "../../frontend")
	http.Handle("/bundle.js", http.FileServer(box))
	http.Handle("/src", http.FileServer(box))

	//For anything else, including index.html and root requests, send back the processed index.html
	http.HandleFunc("/", func(writer http.ResponseWriter, reader *http.Request) {
		templateString, err := box.FindString("./index.html")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		templates := template.Must(template.New("template").Parse(templateString))

		formBytes, err := json.Marshal(config.Form())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		schemaBytes, err := json.Marshal(config.Schema())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		goTemplate := GoTemplate{
			Form:   string(formBytes),
			Schema: string(schemaBytes),
		}
		if err := templates.ExecuteTemplate(writer, "template", goTemplate); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	})

	//Start the web server, set the port to listen to 8080. Without a path it assumes localhost
	listenAddr := fmt.Sprintf("%s:%d", config.Host(), config.Port())
	logrus.Info("Will listen on ", listenAddr)
	err := http.ListenAndServe(listenAddr, nil)
	return err
}
