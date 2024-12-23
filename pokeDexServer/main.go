package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"netcentric/utils"
)

type Pokemon struct {
	Name   string `json:"name"`
	Height int    `json:"height"`
	Weight int    `json:"weight"`
	Types  []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	Abilities []struct {
		Ability struct {
			Name string `json:"name"`
		} `json:"ability"`
	} `json:"abilities"`
	Stats []struct {
		Stat struct {
			Name string `json:"name"`
		} `json:"stat"`
		BaseStat int `json:"base_stat"`
	} `json:"stats"`
}

// Function to read Pokémon data from a JSON file by number
func readPokemonData(number string) (Pokemon, error) {
	filename := fmt.Sprintf("../monsterData/pokemon_data/%s.json", number)

	log.Printf("Attempting to open file: %s", filename)

	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Error: Pokémon data not found for file %s", number)
			return Pokemon{}, fmt.Errorf("pokémon data not found for file %s", number)
		}
		log.Printf("Error: Failed to open file %s: %v", filename, err)
		return Pokemon{}, fmt.Errorf("failed to open file %s: %v", filename, err)
	}
	defer file.Close()

	var pokemon Pokemon
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("Error: Failed to read data from file %s: %v", filename, err)
		return Pokemon{}, fmt.Errorf("failed to read data from file %s: %v", filename, err)
	}

	if err := json.Unmarshal(data, &pokemon); err != nil {
		log.Printf("Error: Failed to unmarshal JSON for file %s: %v", filename, err)
		return Pokemon{}, fmt.Errorf("failed to unmarshal JSON for file %s: %v", filename, err)
	}

	log.Printf("Successfully read data for Pokémon %s", pokemon.Name)
	return pokemon, nil
}

// Function to map Pokémon name to its corresponding number using PokeMap
func getPokemonNumberByName(name string) (string, error) {
	name = strings.Title(strings.ToLower(name))

	// Log lookup attempt
	log.Printf("Looking up Pokémon: %s", name)

	if number, exists := utils.PokeMap[name]; exists {
		log.Printf("Found Pokémon %s with number %s", name, number)
		return number, nil
	}

	log.Printf("Error: Pokémon %s not found in PokeMap", name)
	return "", fmt.Errorf("Pokemon %s not found", name)
}

// Handler for Pokémon requests
func handlePokemonRequest(w http.ResponseWriter, r *http.Request) {
	// Log incoming request
	log.Printf("Received %s request for %s", r.Method, r.URL.Path)

	// Ensure method is GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Printf("Method not allowed: %s", r.Method)
		return
	}

	// Extract 'name' query parameter
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing Pokemon name", http.StatusBadRequest)
		log.Printf("Error: Missing Pokemon name in request")
		return
	}

	// Get the Pokémon number from the name
	number, err := getPokemonNumberByName(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %v", err), http.StatusInternalServerError)
		log.Printf("Error fetching data for %s: %v", name, err)
		return
	}

	// Fetch the Pokémon data from file by number
	pokemon, err := readPokemonData(number)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %v", err), http.StatusInternalServerError)
		log.Printf("Error fetching data for %s: %v", name, err)
		return
	}

	// Respond with the Pokémon data as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pokemon); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		log.Printf("Error encoding response for %s: %v", name, err)
	}
	log.Printf("Successfully responded with data for %s", name)
}

// Main function to start the server
func main() {
	// Handle requests at '/pokemon'
	http.HandleFunc("/pokemon", handlePokemonRequest)

	// Start the server
	log.Printf("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
