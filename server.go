package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/parnurzeal/gorequest"
)

func main() {
	router := httprouter.New()
	godotenv.Load()
	router.GET("/", index)
	router.GET("/:backLink", linkRedirecter)
	router.POST("/", shortener)
	port := os.Getenv("PORT")
	log.Print("Listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fp := filepath.Join("pages", "index.html")
	tmpl, err := template.ParseFiles(fp)
	err = tmpl.ExecuteTemplate(w, "index", nil)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func linkRedirecter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	airtableAPIKey := os.Getenv("APIKEY")
	backLink := ps.ByName("backLink")
	request := gorequest.New()
	getLink := `https://api.airtable.com/v0/appwCEZWtrWK1mITO/links?filterByFormula=SEARCH("` + backLink + `",backLink)`
	resp, body, errs := request.Get(getLink).
		Set("Authorization", airtableAPIKey).
		End()
	if errs != nil {
		log.Fatal(errs)
		fmt.Println(resp)
	}
	if strings.Contains(body, "longURL") {
		splitted := strings.Split(body, `longURL":"`)[1]
		longURL := strings.Split(splitted, `","backLink`)[0]
		http.Redirect(w, r, longURL, 301)
	} else {
		http.Error(w, http.StatusText(404), 404)
	}

}

func shortener(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	airtableAPIKey := os.Getenv("APIKEY")
	recaptchaSecretKey := os.Getenv("SECRETKEY")
	r.ParseForm()
	longLink := strings.Split(r.FormValue("long-url"), "xc0000b6070 0}")[0]
	backLink := strings.Split(r.FormValue("shortened-backlink"), "xc0000b8070 0}")[0]
	recaptcha := strings.Split(r.FormValue("g-recaptcha-response"), "xc0000b8070 0}")[0]
	if longLink == "" {
		fp := filepath.Join("pages", "index.html")
		tmpl, err := template.ParseFiles(fp)
		err = tmpl.ExecuteTemplate(w, "index", "fill all the fields")
		if err != nil {
			log.Println(err.Error())
			http.Error(w, http.StatusText(500), 500)
		}
	} else {
		if backLink == "" {
			fp := filepath.Join("pages", "index.html")
			tmpl, err := template.ParseFiles(fp)
			err = tmpl.ExecuteTemplate(w, "index", "fill all the fields")
			if err != nil {
				log.Println(err.Error())
				http.Error(w, http.StatusText(500), 500)
			}
		} else {
			fmt.Println(longLink)
			fmt.Println(backLink)
			postLink := `https://www.google.com/recaptcha/api/siteverify?secret=` + recaptchaSecretKey + `&response=` + recaptcha
			request := gorequest.New()
			resp, body, errs := request.Post(postLink).End()
			if errs != nil {
				fmt.Println(resp)
				log.Fatal(errs)
			}
			splitted := strings.Split(body, `"success":`)[1]
			recaptchaResponse := strings.Split(splitted, ",")[0]
			if recaptchaResponse == " true" {
				if strings.Contains(backLink, "/") {
					fp := filepath.Join("pages", "index.html")
					tmpl, err := template.ParseFiles(fp)
					err = tmpl.ExecuteTemplate(w, "index", "invalid backlink")
					if err != nil {
						log.Println(err.Error())
						http.Error(w, http.StatusText(500), 500)
					}
				} else {

					request := gorequest.New()
					getLink := `https://api.airtable.com/v0/appwCEZWtrWK1mITO/links?filterByFormula=SEARCH("` + backLink + `",backLink)`
					resp, body, errs := request.Get(getLink).
						Set("Authorization", airtableAPIKey).
						End()
					if errs != nil {
						log.Fatal(errs)
						fmt.Println(resp)
					}
					if strings.Contains(body, "longURL") {
						fp := filepath.Join("pages", "index.html")
						tmpl, err := template.ParseFiles(fp)
						err = tmpl.ExecuteTemplate(w, "index", "backlink already exists")
						if err != nil {
							log.Println(err.Error())
							http.Error(w, http.StatusText(500), 500)
						}
					} else {
						request := gorequest.New()
						postBody := `{"fields":{"longURL":"` + longLink + `","backLink":"` + backLink + `"}}`
						resp, body, errs := request.Post("https://api.airtable.com/v0/appwCEZWtrWK1mITO/links").
							Set("Authorization", airtableAPIKey).
							Send(postBody).
							End()
						if errs != nil {
							log.Fatal(errs)
							fmt.Println(resp)
						}
						fmt.Println(body)
						fp := filepath.Join("pages", "index.html")
						tmpl, err := template.ParseFiles(fp)
						err = tmpl.ExecuteTemplate(w, "index", "successfully shortened")
						if err != nil {
							log.Println(err.Error())
							http.Error(w, http.StatusText(500), 500)
						}
					}
				}
			} else {
				fp := filepath.Join("pages", "index.html")
				tmpl, err := template.ParseFiles(fp)
				err = tmpl.ExecuteTemplate(w, "index", "recaptcha failed")
				if err != nil {
					log.Println(err.Error())
					http.Error(w, http.StatusText(500), 500)
				}
			}
		}
	}
}
