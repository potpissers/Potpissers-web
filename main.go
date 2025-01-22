package main

import (
	"html/template"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		template, err := template.ParseFiles("main.html", "main-home.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		template.ExecuteTemplate(w, "main.html", nil)
	})
	// http.HandleFunc("hcf", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "hcf.html")
	// })
	// http.HandleFunc("mz", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "mz.html")
	// })

	http.Handle("/style.css", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	http.Handle("/potpisser.jpg", http.StripPrefix("/", http.FileServer(http.Dir("."))))

	http.ListenAndServe(":8080", nil)
}
