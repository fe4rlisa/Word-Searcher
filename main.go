package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Page struct {
	Title string
}

type Definition struct {
	Word     string    `json:"word"`
	Meanings []Meaning `json:"meanings"`
}

type Meaning struct {
	PartOfSpeech string             `json:"partOfSpeech"`
	Definitions  []DefinitionDetail `json:"definitions"`
}

type DefinitionDetail struct {
	Definition string `json:"definition"`
}

var tmpl *template.Template

func mainHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{Title: "Word Searcher"}
	tmpl.ExecuteTemplate(w, "index.html", p)
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	wordDef := make(map[string]string)
	r.ParseForm()
	words := r.PostFormValue("words")
	wordsSlice := strings.Split(words, "\n")

	orderedDefinitions := make([]struct {
		Word       string
		Definition string
	}, len(wordsSlice))

	for i, word := range wordsSlice {
		//wordsSlice[i] = strings.TrimSpace(word)
		//wordDef[wordsSlice[i]] = define(wordsSlice[i])
		trimmedWord := strings.TrimSpace(word)
		definition := define(trimmedWord)
		orderedDefinitions[i] = struct {
			Word       string
			Definition string
		}{Word: trimmedWord, Definition: definition}
	}

	log.Println("Words: ", wordsSlice)
	for word, value := range wordDef {
		log.Printf("Word: %s, Definition: %s\n", word, value)
	}

	var responseHTML strings.Builder
	for _, entry := range orderedDefinitions {
		responseHTML.WriteString("<div class='definition'>")
		responseHTML.WriteString("<h2>" + entry.Word + "</h2>")
		responseHTML.WriteString("<p>" + entry.Definition + "</p>")
		responseHTML.WriteString("</div>")
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(responseHTML.String()))

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func define(words string) string {
	word := words
	var d []Definition

	data, err := http.Get("https://api.dictionaryapi.dev/api/v2/entries/en/" + word)
	if err != nil {
		log.Println("Error fetching data", err)
		return ""
	}
	defer data.Body.Close()

	err = json.NewDecoder(data.Body).Decode(&d)
	if err != nil {
		log.Println("Error decodig JSON", err)
		return ""
	}

	var definitions []string
	for _, def := range d {
		for _, meaning := range def.Meanings {
			for _, definition := range meaning.Definitions {
				definitions = append(definitions, definition.Definition)
				//log.Println(definition.Definition)
			}
		}
	}

	return strings.Join(definitions, "; ")
}

func main() {

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	tmpl = template.Must(template.ParseGlob("tmpl/*.html"))

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Get(("/"), mainHandler)
	r.Post(("/define"), formHandler)

	http.ListenAndServe(":8080", r)

}
