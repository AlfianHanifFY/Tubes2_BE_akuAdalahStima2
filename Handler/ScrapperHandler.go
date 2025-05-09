package handler

import (
	"fmt"
	"net/http"
	"stima-2-be/Element"
	"stima-2-be/scrapper"
)

func ScrapHandler(w http.ResponseWriter, r *http.Request) {
	scrapper.Scrapper()
	fmt.Fprintln(w, "Scraping selesai.")
	Element.LoadElementsFromFile("output.json")
	fmt.Fprintln(w, "Load selesai")
}
