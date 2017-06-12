package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
)

type IndexPage struct {
	Photos  []string
	LogedIn bool
}

type LoginPage struct {
	Body      string
	FirstName string
	LastName  string
	Email     string
	Error     string
}
type UploadPage struct {
	Error string
	Msg   string
}

func getPhotos() []string {
	photos := make([]string, 0)
	filepath.Walk("assets/img", func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}
		path = strings.Replace(path, "\\", "/", -1)
		photos = append(photos, path)
		return nil
	})
	return photos
}

var store = sessions.NewCookieStore([]byte("HelloWorld"))

func loginPage(res http.ResponseWriter, req *http.Request) {
	loginError := ""
	session, _ := store.Get(req, "session")
	str, _ := session.Values["logged-in"].(string)
	if str == "YES" {
		http.Redirect(res, req, "/admin", 302)
		return
	}
	if req.Method == "POST" {
		email := req.FormValue("email")
		password := req.FormValue("password")
		if email == "test@example.com" && password == "test" {
			session.Values["logged-in"] = "YES"
			session.Save(req, res)
			http.Redirect(res, req, "/admin", 302)
			return
		} else {
			loginError = "Invalid Credential. Please Resubmit"
		}
	}
	tpl, err := template.ParseFiles("assets/tpl/login.gohtml", "assets/tpl/header.gohtml")
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}
	err = tpl.Execute(res, LoginPage{
		Error: loginError,
	})
}

func admin(res http.ResponseWriter, req *http.Request) {
	uploadError := ""
	successMsg := ""
	session, _ := store.Get(req, "session")
	str, _ := session.Values["logged-in"].(string)
	if str != "YES" {
		http.Redirect(res, req, "/login", 302)
		return
	}
	if req.Method == "POST" {
		// <input type="file" name="file">
		src, hdr, err := req.FormFile("file")
		if err != nil {
			http.Error(res, "Invalid File.", 500)
			return
		}

		defer src.Close()
		// create a new file
		// make sure you have a "tmp" directory in your web root
		dst, err := os.Create("assets/img/" + hdr.Filename)
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}

		defer dst.Close()

		// copy the uploaded file into the new file
		io.Copy(dst, src)
	}
	tpl, err := template.ParseFiles("assets/tpl/admin.gohtml", "assets/tpl/header.gohtml")
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}
	err = tpl.Execute(res, UploadPage{
		Error: uploadError,
		Msg:   successMsg,
	})
	if err != nil {
		http.Error(res, err.Error(), 500)
	}
}

func index(res http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "session")
	str, _ := session.Values["logged-in"].(string)
	logged := false
	if str == "YES" {
		logged = true
	}

	tpl, err := template.ParseFiles("assets/tpl/index.gohtml", "assets/tpl/header.gohtml")
	if err != nil {
		fmt.Println(err)
		http.Error(res, err.Error(), 500)
		return
	}
	err = tpl.Execute(res, IndexPage{
		Photos:  getPhotos(),
		LogedIn: logged,
	})
	if err != nil {
		fmt.Println(err)
		http.Error(res, err.Error(), 500)
	}
}

func logout(res http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "session")
	str, _ := session.Values["logged-in"].(string)
	if str == "YES" {
		delete(session.Values, "logged-in")
		session.Save(req, res)
		http.Redirect(res, req, "/", 302)
	} else {
		http.Redirect(res, req, "/login", 302)
	}
}

func deletePic(res http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "session")
	str, _ := session.Values["logged-in"].(string)
	if str != "YES" {
		http.Redirect(res, req, "/", 302)
		return
	}
	if req.Method == "POST" {
		imgName := req.FormValue("imgName")
		err := os.Remove(imgName)
		if err != nil {
			http.Error(res, err.Error(), 500)
		}
	}

	tpl, err := template.ParseFiles("assets/tpl/delete.gohtml", "assets/tpl/header.gohtml")
	if err != nil {
		http.Error(res, err.Error(), 500)
	}
	err = tpl.Execute(res, IndexPage{
		Photos: getPhotos(),
	})
	if err != nil {
		http.Error(res, err.Error(), 500)
	}
}

func main() {
	http.HandleFunc("/delete", deletePic)
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))
	http.HandleFunc("/", index)
	http.HandleFunc("/admin", admin)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/logout", logout)
	http.ListenAndServe(":80", context.ClearHandler(http.DefaultServeMux))
}
