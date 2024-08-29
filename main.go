package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

var (
	docketCounter   = 1
	logsheetCounter = 1
	mutex           sync.Mutex
	dockets         = make(map[string]Docket) // Store dockets by OrderNo
)

// Docket is made for every Delivery Order (DO)
type Docket struct {
	OrderNo       string  `json:"OrderNo"`
	Customer      string  `json:"Customer"`
	PickUpPoint   string  `json:"PickUpPoint"`
	DeliveryPoint string  `json:"DeliveryPoint"`
	Quantity      int     `json:"Quantity"`
	Volume        float64 `json:"Volume"`
	Status        string  `json:"Status"`
	TruckNo       string  `json:"TruckNo"`
	LogsheetNo    string  `json:"LogsheetNo"`
}

// Logsheet represent the docket(s) and the truck assigned to deliver the DOs
type LogsheetRequest struct {
	Dockets []string `json:"Dockets"`
	TruckNo string   `json:"TruckNo"`
}

// main sets up the HTTP server and routes
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/docket", createDocketHandler).Methods("POST")
	r.HandleFunc("/docket/{orderNo:[A-Za-z0-9]+}", getDocketHandler).Methods("GET")
	r.HandleFunc("/dockets", getAllDocketsHandler).Methods("GET")
	r.HandleFunc("/logsheet", createLogsheetHandler).Methods("POST")
	r.HandleFunc("/logsheet/{logsheetNo:[A-Za-z0-9]+}", getLogsheetHandler).Methods("GET")

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// this function return a welcome message at root URL ("/")
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to the Transport Management System (TMS) API"))
}

// QUESTION 1: this function handles POST requests to "/docket" and creates a new docket
func createDocketHandler(w http.ResponseWriter, r *http.Request) {
	var docket Docket
	if err := json.NewDecoder(r.Body).Decode(&docket); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	docket.OrderNo = fmt.Sprintf("TDN%04d", docketCounter)
	docket.Status = "Created"
	dockets[docket.OrderNo] = docket // Store docket in the map
	docketCounter++
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docket)
}

// QUESTION 2: this function handles GET requests to "/docket/{orderNo}" and returns a specific docket by order number
func getDocketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderNo := vars["orderNo"]

	mutex.Lock()
	docket, exists := dockets[orderNo]
	mutex.Unlock()

	if !exists {
		http.Error(w, "Docket not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docket)
}

// QUESTION 3: this function handles GET requests to "/dockets" and returns a list of all dockets
func getAllDocketsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	mutex.Lock()
	var docketList []Docket
	for _, docket := range dockets {
		docketList = append(docketList, docket)
	}
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docketList)
}

// QUESTION 4: this function handles POST requests to "/logsheet" and creates a new logsheet, updating dockets accordingly
func createLogsheetHandler(w http.ResponseWriter, r *http.Request) {
	var request LogsheetRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	logsheetNo := fmt.Sprintf("DT%04d", logsheetCounter)
	logsheetCounter++

	updatedDockets := []Docket{}
	for _, orderNo := range request.Dockets {
		if docket, exists := dockets[orderNo]; exists {
			docket.TruckNo = request.TruckNo
			docket.LogsheetNo = logsheetNo
			dockets[orderNo] = docket
			updatedDockets = append(updatedDockets, docket)
		}
	}
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedDockets)
}

// QUESTION 5: this function handles GET requests to "/logsheet/{logsheetNo}" and returns all dockets associated with a specific logsheet number
func getLogsheetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	logsheetNo := vars["logsheetNo"]

	mutex.Lock()
	var logsheetDockets []Docket
	for _, docket := range dockets {
		if docket.LogsheetNo == logsheetNo {
			logsheetDockets = append(logsheetDockets, docket)
		}
	}
	mutex.Unlock()

	if len(logsheetDockets) == 0 {
		http.Error(w, "Logsheet not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logsheetDockets)
}
