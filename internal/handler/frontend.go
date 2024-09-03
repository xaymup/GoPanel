package handler

import (
    "net/http"
	"github.com/CloudyKit/jet/v6"
    "gopanel/internal/util"
	"gopanel/web"
	"gopanel/cmd"
	"log"
	"fmt"
	"strings"
)


// var views = web.GetSet()

var developmentMode = cmd.GetMode()

func FrontendHandler(w http.ResponseWriter, r *http.Request) {

	var templateName string
	if util.CheckStack() {
	// if stackStatus{ 
		session, _ := Store.Get(r, "session")
		
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

			parts := strings.Split(path, "/")

			if parts[0] == "" {
				path = "/home" // default template (e.g., / -> index.jet)
			} else {
				path = parts[0]
			}

			templateName = fmt.Sprintf("/public/%s.jet", path)
		}
	} else {
		templateName = "public/install"
	}

	var views *jet.Set

	if developmentMode {
		views = jet.NewSet(
			jet.NewOSFileSystemLoader("./web"), // Load templates from the "templates" directory
			jet.InDevelopmentMode(),                  // Disable caching for development mode
		)
    } else {
		views = web.GetViews()
    }
	
	tmpl, err := views.GetTemplate(templateName)
    if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
        http.Redirect(w, r, "/", http.StatusFound)
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

