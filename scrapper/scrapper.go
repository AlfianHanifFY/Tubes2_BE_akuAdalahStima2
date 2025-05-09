package scrapper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"stima-2-be/Element"
	"strings"
)

func Scrapper() {
	url := "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		panic(err)
	}
	html := buf.String()

	var allElements []Element.Element

	// --- Tangani Starting Elements ---
	startingRegex := regexp.MustCompile(`(?is)<span class="mw-headline" id="Starting_elements">.*?</span>.*?(<table.*?>.*?</table>)`)
	startingMatch := startingRegex.FindStringSubmatch(html)
	if len(startingMatch) > 1 {
		tableHtml := startingMatch[1]
		elements := extractElementsFromTable(tableHtml, "0")
		allElements = append(allElements, elements...)
	}

	// --- Tangani Tier Elements ---
	tierTableRegex := regexp.MustCompile(`(?is)<span class="mw-headline" id="(Tier_\d+)_elements">.*?</span>.*?(<table.*?>.*?</table>)`)
	tierTableMatches := tierTableRegex.FindAllStringSubmatch(html, -1)

	for _, match := range tierTableMatches {
		tierRaw := match[1] // e.g., Tier_1
		tier := strings.Split(tierRaw, "_")[1]
		tableHtml := match[2]

		elements := extractElementsFromTable(tableHtml, tier)
		allElements = append(allElements, elements...)
	}

	// --- Tangani Unlocking Requirement Table ---
	unlockingRegex := regexp.MustCompile(`(?is)<span class="mw-headline" id="Element">.*?</span>.*?(<table.*?>.*?</table>)`)
	unlockingMatch := unlockingRegex.FindStringSubmatch(html)
	if len(unlockingMatch) > 1 {
		tableHtml := unlockingMatch[1]
		elements := extractElementsFromTable(tableHtml, "-1") // Gunakan tier khusus misal "-1" atau "unlock"
		allElements = append(allElements, elements...)
	}

	// --- Tambahkan Time dengan Tier 0 secara manual ---
	allElements = append(allElements, Element.Element{
		Root:  "Time",
		Left:  "",
		Right: "",
		Tier:  "0", // Time selalu di-set ke tier 0
	})

	// Simpan sebagai JSON
	jsonData, err := json.MarshalIndent(allElements, "", "  ")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("output.json", jsonData, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Berhasil disimpan di output.json")
}

func extractElementsFromTable(tableHtml string, tier string) []Element.Element {
	var elements []Element.Element

	rowRegex := regexp.MustCompile(`(?is)<tr.*?>.*?</tr>`)
	rows := rowRegex.FindAllString(tableHtml, -1)

	for _, row := range rows {
		tdRegex := regexp.MustCompile(`(?is)<td.*?>.*?</td>`)
		tds := tdRegex.FindAllString(row, -1)
		if len(tds) < 1 {
			continue
		}

		// Ambil nama elemen di kolom kiri (Root)
		elementRegex := regexp.MustCompile(`(?is)<a href="/wiki/[^"]*" title="[^"]*">(.*?)</a>`)
		elementMatch := elementRegex.FindStringSubmatch(tds[0])
		root := "UNKNOWN"
		if len(elementMatch) > 1 {
			root = cleanText(elementMatch[1])
		}

		// Cek apakah ada komposer di kolom kanan
		if len(tds) >= 2 {
			komposers := extractKomposers(tds[1])

			// Jika ada kombinasi (pasangan), simpan sebagai kombinasi
			if len(komposers) > 0 {
				for _, pair := range komposers {
					elements = append(elements, Element.Element{
						Root:  root,
						Left:  pair[0],
						Right: pair[1],
						Tier:  tier,
					})
				}
				continue
			}

			// Jika tidak ada pasangan, tapi masih dalam daftar unlock seperti "Time"
			// maka asumsikan elemen mandiri, tier 0
			elements = append(elements, Element.Element{
				Root:  root,
				Left:  "",
				Right: "",
				Tier:  "0",
			})
			continue
		}

		// Jika tidak ada kolom kanan, simpan sebagai elemen mandiri
		elements = append(elements, Element.Element{
			Root:  root,
			Left:  "",
			Right: "",
			Tier:  tier,
		})
	}

	return elements
}

// Bersihkan teks
func cleanText(s string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

// Ambil 2 komposer dari <li> dan abaikan <span>
func extractKomposers(td string) [][]string {
	liRegex := regexp.MustCompile(`(?is)<li[^>]*>.*?</li>`)
	liMatches := liRegex.FindAllString(td, -1)

	var pairs [][]string

	for _, li := range liMatches {
		spanRegex := regexp.MustCompile(`(?is)<span[^>]*>.*?</span>`)
		cleanLi := spanRegex.ReplaceAllString(li, "")

		aRegex := regexp.MustCompile(`(?is)<a[^>]*>(.*?)</a>`)
		aMatches := aRegex.FindAllStringSubmatch(cleanLi, -1)

		if len(aMatches) == 2 {
			left := cleanText(aMatches[0][1])
			right := cleanText(aMatches[1][1])
			pairs = append(pairs, []string{left, right})
		}
	}

	return pairs
}
