package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Pokemon structure to hold detailed data
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

// PokemonList represents a paginated list of Pokémon from the API
type PokemonList struct {
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
}

// Fetch Pokémon data from the PokéAPI
func fetchPokemonData(name string) (Pokemon, error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", name)
	resp, err := http.Get(url)
	if err != nil {
		return Pokemon{}, fmt.Errorf("failed to fetch data from API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Pokemon{}, fmt.Errorf("failed to fetch data: status code %d", resp.StatusCode)
	}

	var pokemon Pokemon
	if err := json.NewDecoder(resp.Body).Decode(&pokemon); err != nil {
		return Pokemon{}, fmt.Errorf("failed to decode API response: %v", err)
	}
	return pokemon, nil
}

// Save Pokémon data to a separate JSON file inside a folder, named by number
func savePokemonDataToFile(pokemon Pokemon, index int, folder string) {
	// Ensure folder exists
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		log.Printf("failed to create folder: %v", err)
		return
	}

	// Save each Pokémon in the folder with a numbered filename
	filename := filepath.Join(folder, fmt.Sprintf("%d.json", index))
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("failed to create JSON file for %s: %v", pokemon.Name, err)
		return
	}
	defer file.Close()

	data, err := json.MarshalIndent(pokemon, "", "  ")
	if err != nil {
		log.Printf("failed to marshal JSON data for %s: %v", pokemon.Name, err)
		return
	}

	_, err = file.Write(data)
	if err != nil {
		log.Printf("failed to write to JSON file for %s: %v", pokemon.Name, err)
		return
	}

	log.Printf("Pokemon %s data saved to %s", pokemon.Name, filename)
}

// Fetch a list of all Pokémon from the PokéAPI
func fetchAllPokemonNames() ([]string, error) {
	url := "https://pokeapi.co/api/v2/pokemon?limit=1000" // Adjust the limit as needed
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Pokémon list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: status code %d", resp.StatusCode)
	}

	var pokemonList PokemonList
	if err := json.NewDecoder(resp.Body).Decode(&pokemonList); err != nil {
		return nil, fmt.Errorf("failed to decode Pokémon list: %v", err)
	}

	var names []string
	for _, p := range pokemonList.Results {
		names = append(names, p.Name)
	}

	return names, nil
}

func main() {
	// Define the folder where Pokémon data will be stored
	folder := "pokemon_data"

	names, err := fetchAllPokemonNames()
	if err != nil {
		log.Fatalf("Failed to fetch all Pokémon names: %v", err)
	}

	for index, name := range names {
		// Fetch full data for each Pokémon
		pokemon, err := fetchPokemonData(name)
		if err != nil {
			log.Printf("Failed to fetch data for %s: %v", name, err)
			continue
		}

		// Save the fetched Pokémon data to a separate file inside the folder (named by number)
		savePokemonDataToFile(pokemon, index+1, folder)
	}
}
