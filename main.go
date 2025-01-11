package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
)

func fetchMetarInfo(icaoCode string) (string, error) {
	// Instantiate the collector
	c := colly.NewCollector()

	// Result variable to hold METAR data
	var metarInfo string

	// OnHTML callback to process the page content
	c.OnHTML("#div_metar tr.aeroTRMetar.aeroTRMetar.celdaColoreada", func(e *colly.HTMLElement) {
		// Iterate over each <td> element inside the current <tr>
		e.ForEach("td", func(_ int, td *colly.HTMLElement) {
			// Directly check if <td> content starts with ICAO code
			if td.Text[:4] == icaoCode {
				// Extract and return the METAR information from the <td>
				metarInfo = td.Text
			}
		})
	})

	// OnError callback to handle errors more efficiently
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Error accessing", r.Request.URL, "-", err)
	})

	// Visit the target URL
	err := c.Visit("https://www.inumet.gub.uy/aeronautica/productos-aeronauticos")
	if err != nil {
		return "", err
	}

	// Return METAR info or error
	if metarInfo == "" {
		return "", fmt.Errorf("METAR info not found")
	}
	return metarInfo, nil
}

func metarHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ICAO code from query parameter
	icaoCode := r.URL.Query().Get("airport")
	if icaoCode == "" {
		http.Error(w, "Missing ICAO code", http.StatusBadRequest)
		return
	}

	// Start timer for execution time
	startTime := time.Now()

	// Fetch METAR info
	metarInfo, err := fetchMetarInfo(icaoCode)
	if err != nil {
		http.Error(w, "Error fetching METAR info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate execution time
	duration := time.Since(startTime)

	// Send the METAR info and execution time in the response
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"icaoCode": "%s", "metarInfo": "%s", "executionTime": "%v"}`, icaoCode, metarInfo, duration)
}

func main() {
	// Setup HTTP server and endpoint
	http.HandleFunc("/get_metar", metarHandler)

	// Start the server
	port := ":8080"
	fmt.Println("Server running on http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
