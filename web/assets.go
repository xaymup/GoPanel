package web

import (
    "embed"
    "log"
	"github.com/CloudyKit/jet/v6"
	"io/fs"
)

//go:embed *
var content embed.FS

var initialzed bool

var Loader = jet.NewInMemLoader()

var views = jet.NewSet(
	Loader,
	jet.InDevelopmentMode(),              // Use in development mode to auto-reload templates
)

func readFS(){
	err := fs.WalkDir(content, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        
        // Skip directories
        if d.IsDir() {
            return nil
        }

        // Read the file content
        data, err := content.ReadFile(path)
        if err != nil {
            return err
        }
		
		Loader.Set(path,string(data))

		initialzed = true
        
        return nil
    })


    if err != nil {
        log.Fatal(err)
    }
}



func GetViews() (*jet.Set) {
	if !initialzed {
		readFS()
	}
	return views
}