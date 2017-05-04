package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/stts-se/rbg2p"
)

type g2pMutex struct {
	g2ps  map[string]rbg2p.RuleSet
	mutex *sync.RWMutex
}

var g2p = g2pMutex{
	g2ps:  make(map[string]rbg2p.RuleSet),
	mutex: &sync.RWMutex{},
}

func g2pMain_Handler(w http.ResponseWriter, r *http.Request) {
	// TODO error if file not found
	http.ServeFile(w, r, "./src/g2p_demo.html")
}

var wSplitRe = regexp.MustCompile(" *, *")

// Word internal struct for json
type Word struct {
	Orth    string   `json:"orth"`
	Transes []string `json:"transes"`
}

func transcribe(lang string, word string) (Word, error, int) {
	g2p.mutex.RLock()
	defer g2p.mutex.RUnlock()
	ruleSet, ok := g2p.g2ps[lang]
	if !ok {
		msg := "unknown 'lang': " + lang
		langs := listLanguages()
		msg = fmt.Sprintf("%s. Known 'lang' values: %s", msg, strings.Join(langs, ", "))
		return Word{}, fmt.Errorf(msg), http.StatusBadRequest
	}

	transes, err := ruleSet.Apply(word)
	if err != nil {
		msg := fmt.Sprintf("couldn't transcribe word : %v", err)
		return Word{}, fmt.Errorf(msg), http.StatusInternalServerError
	}
	tRes := []string{}
	for _, trans := range transes {
		tRes = append(tRes, trans.String(ruleSet.PhonemeDelimiter))
	}
	res := Word{word, tRes}
	return res, nil, http.StatusOK
}

func transcribe_Handler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	lang := vars["lang"]
	if "" == lang {
		msg := "no value for the expected 'lang' parameter"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	word := vars["word"]
	if "" == word {
		msg := "no value for the expected 'word' parameter"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	word = strings.ToLower(word)

	res, err, status := transcribe(lang, word)
	if err != nil {
		log.Print("%s\n", err)
		http.Error(w, fmt.Sprintf("%s", err), status)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	j, err := json.Marshal(res)
	if err != nil {
		msg := fmt.Sprintf("failed json marshalling : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(j))
}

func transcribe_OnlyFirstTrans_Handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]
	if "" == lang {
		msg := "no value for the expected 'lang' parameter"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	word := vars["word"]
	if "" == word {
		msg := "no value for the expected 'word' parameter"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	word = strings.ToLower(word)
	res, err, status := transcribe(lang, word)
	if err != nil {
		log.Print("%s\n", err)
		http.Error(w, fmt.Sprintf("%s", err), status)
		return
	}

	fmt.Fprintf(w, string(res.Transes[0]))
}

func listLanguages() []string {
	var res []string
	for name := range g2p.g2ps {
		res = append(res, name)
	}
	return res
}

func list_Handler(w http.ResponseWriter, r *http.Request) {
	g2p.mutex.RLock()
	res := listLanguages()
	g2p.mutex.RUnlock()

	sort.Strings(res)
	j, err := json.Marshal(res)
	if err != nil {
		msg := fmt.Sprintf("failed json marshalling : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(j))
}

// langFromFilePath returns the base file name stripped from any '.g2p' extension
func langFromFilePath(p string) string {
	b := filepath.Base(p)
	if strings.HasSuffix(b, ".g2p") {
		b = b[0 : len(b)-4]
	}
	return b
}

func main() {

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "g2pserver <G2P FILES DIR>\n")
		os.Exit(0)
	}

	// g2p file dir. Each file in dir with .g2p extension
	// is treated as a g2p file
	var dir = os.Args[1]

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(0)
	}

	// populate map of g2p rules from files.
	// The base file name minus '.g2p' is the language name.
	var fn string
	for _, f := range files {
		fn = filepath.Join(dir, f.Name())
		if !strings.HasSuffix(fn, ".g2p") {
			fmt.Fprintf(os.Stderr, "g2pserver: skipping file: '%s'\n", fn)
			continue
		}

		ruleSet, err := rbg2p.LoadFile(fn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			fmt.Fprintf(os.Stderr, "g2pserver: skipping file: '%s'\n", fn)
			continue
		}

		lang := langFromFilePath(fn)
		g2p.mutex.Lock()
		g2p.g2ps[lang] = ruleSet
		g2p.mutex.Unlock()
		fmt.Fprintf(os.Stderr, "g2p server: loaded file '%s'\n", fn)
	}

	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/rbg2p", g2pMain_Handler) //.Methods("get")
	r.HandleFunc("/", g2pMain_Handler)      //.Methods("get")

	s := r.PathPrefix("/rbg2p").Subrouter()

	s.HandleFunc("/transcribe/{lang}/{word}", transcribe_Handler)
	s.HandleFunc("/list", list_Handler) //.Methods("get", "post")

	// get one trans only
	s = r.PathPrefix("/rbg2p/onetrans").Subrouter()
	s.HandleFunc("/{lang}/{word}", transcribe_OnlyFirstTrans_Handler)

	port := ":6771"
	log.Printf("starting g2p server at port %s\n", port)
	err = http.ListenAndServe(port, r)
	if err != nil {

		log.Fatalf("no fun: %v\n", err)
	}

}
