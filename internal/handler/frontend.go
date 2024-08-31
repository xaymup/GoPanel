package handler

import (
    "net/http"
	"github.com/CloudyKit/jet/v6"
    "gopanel/internal/util"
	"log"
	"fmt"
	"strings"
)



var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./web"), // Load templates from the "views" directory
	jet.InDevelopmentMode(),              // Use in development mode to auto-reload templates
)

func FrontendHandler(w http.ResponseWriter, r *http.Request) {
	var templateName string
	if util.CheckIfStackReady() {
		session, _ := store.Get(r, "session")
		
		// Check if the user is authenticated
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			if !util.FileExists(otpFile) {
				templateName = "public/account"
			} else {
				// fileToServe = "web/public/login.html"
				templateName = "public/login"
			}
		} else {
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "home" // default template (e.g., / -> index.jet)
			}
			templateName = fmt.Sprintf("/public/%s.jet", path)
		}
	} else {
		templateName = "public/install"
	}





	// input := map[string]interface{}{
	// 	"Title": "Home Page",
	// }
	// err := templates.ExecuteTemplate(w, fileToServe, input)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }
	// data, err := ioutil.ReadFile(fileToServe)
	// if err != nil {
	// 	http.Error(w, "File not found", http.StatusNotFound)
	// 	return
	// }
	// w.Header().Set("Content-Type", "text/html")
	// w.Write(data)

	// err := templates.ExecuteTemplate(w, "view.html", map[string]string{
    //     "ContentTemplate": tmpl,
    // })
    // if err != nil {
    //     http.Error(w, err.Error(), http.StatusInternalServerError)
    // }
	// lmao,_ := templates.ParseFiles(tmpl)

	// data := map[string]interface{}{
    //     "ContentTemplate": fmt.Sprintf("{{template \"%s\"}}", tmpl),
    // }
    // err := templates.ExecuteTemplate(w, "view.html", data)
    // if err != nil {
    //     http.Error(w, err.Error(), http.StatusInternalServerError)
    // }

	// Load and render the template

	tmpl, err := views.GetTemplate(templateName)
    if err != nil {
        http.Error(w, fmt.Sprintf("Template not found: %s", err), http.StatusNotFound)
        return
    }

	// Set common data
	data := make(jet.VarMap)
	// data.Set("Title", strings.Title(path) + " Page")

	// Render the template
	if err := tmpl.Execute(w, data, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Println(r.Method, r.URL, r.RemoteAddr)
}

