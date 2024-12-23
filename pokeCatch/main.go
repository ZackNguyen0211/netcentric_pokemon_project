package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
	"netcentric/utils" 
)

const baseURL = "https://pokeapi.co/api/v1"

type Region struct {
	Name      string `json:"name"`
	Locations []struct {
		Name string `json:"name"`
	} `json:"locations"`
}

type Location struct {
	Areas []struct {
		Name string `json:"name"`
	} `json:"areas"`
}

type Pokemon struct {
	Name    string `json:"name"`
	Sprites struct {
		FrontDefault string `json:"front_default"`
	} `json:"sprites"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
}

type Encounter struct {
	Pokemon struct {
		Name string `json:"name"`
	} `json:"pokemon"`
}

type Area struct {
	PokemonEncounters []Encounter `json:"pokemon_encounters"`
}

func get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func getRegions() ([]Region, error) {
	url := fmt.Sprintf("%s/region", baseURL)
	body, err := get(url)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []Region `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

func getRegionDetails(name string) (*Region, error) {
	url := fmt.Sprintf("%s/region/%s", baseURL, name)
	body, err := get(url)
	if err != nil {
		return nil, err
	}

	var region Region
	if err := json.Unmarshal(body, &region); err != nil {
		return nil, err
	}

	return &region, nil
}

func getLocationDetails(name string) (*Location, error) {
	url := fmt.Sprintf("%s/location/%s", baseURL, name)
	body, err := get(url)
	if err != nil {
		return nil, err
	}

	var location Location
	if err := json.Unmarshal(body, &location); err != nil {
		return nil, err
	}

	return &location, nil
}

func getArea(name string) (*Area, error) {
	url := fmt.Sprintf("%s/location-area/%s", baseURL, name)
	body, err := get(url)
	if err != nil {
		return nil, err
	}

	var area Area
	if err := json.Unmarshal(body, &area); err != nil {
		return nil, err
	}

	return &area, nil
}

func getPokemon(name string) (*Pokemon, error) {
	pokeID, exists := utils.PokeMap[strings.Title(name)]
	if !exists {
		return nil, fmt.Errorf("pokémon not found in map: %s", name)
	}
	url := fmt.Sprintf("%s/pokemon/%s", baseURL, pokeID)
	body, err := get(url)
	if err != nil {
		return nil, err
	}

	var pokemon Pokemon
	if err := json.Unmarshal(body, &pokemon); err != nil {
		return nil, err
	}

	return &pokemon, nil
}

func getRandomEncounter(possible []Encounter) (*Pokemon, error) {
	if len(possible) == 0 {
		return nil, fmt.Errorf("no possible encounters")
	}

	rand.Seed(time.Now().UnixNano())
	selected := possible[rand.Intn(len(possible))].Pokemon.Name
	return getPokemon(selected)
}

func downloadImage(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func openImage(filepath string) error {
	cmd := exec.Command("cmd", "/c", "start", filepath)
	return cmd.Start()
}

func main() {
	// Step 1: Fetch and list all regions
	regions, err := getRegions()
	if err != nil {
		fmt.Printf("Error fetching regions: %v\n", err)
		return
	}

	// For simplicity, let's select the first region
	selectedRegion := regions[0].Name
	fmt.Printf("Selected Region: %s\n", selectedRegion)

	// Step 2: Fetch and list locations within the selected region
	regionDetails, err := getRegionDetails(selectedRegion)
	if err != nil {
		fmt.Printf("Error fetching region details: %v\n", err)
		return
	}

	// For simplicity, let's select the first location
	if len(regionDetails.Locations) == 0 {
		fmt.Println("No locations found in this region.")
		return
	}
	selectedLocation := regionDetails.Locations[0].Name
	fmt.Printf("Selected Location: %s\n", selectedLocation)

	// Step 3: Fetch and list areas within the selected location
	locationDetails, err := getLocationDetails(selectedLocation)
	if err != nil {
		fmt.Printf("Error fetching location details: %v\n", err)
		return
	}

	// For simplicity, let's select the first area
	if len(locationDetails.Areas) == 0 {
		fmt.Println("No areas found in this location.")
		return
	}
	selectedArea := locationDetails.Areas[0].Name
	fmt.Printf("Selected Area: %s\n", selectedArea)

	// Step 4: Fetch and display a random Pokémon encounter from the selected area
	area, err := getArea(selectedArea)
	if err != nil {
		fmt.Printf("Error fetching area: %v\n", err)
		return
	}

	encounteredPokemon, err := getRandomEncounter(area.PokemonEncounters)
	if err != nil {
		fmt.Printf("Error fetching encounter: %v\n", err)
		return
	}

	fmt.Printf("You encountered a %s!\n", encounteredPokemon.Name)
	fmt.Printf("Image: %s\n", encounteredPokemon.Sprites.FrontDefault)
	fmt.Println("Stats:")
	for _, stat := range encounteredPokemon.Stats {
		fmt.Printf("- %s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	// Step 5: Download and open the Pokémon image
	imageURL := encounteredPokemon.Sprites.FrontDefault
	filepath := "pokemon_image.png"
	err = downloadImage(imageURL, filepath)
	if err != nil {
		fmt.Printf("Error downloading image: %v\n", err)
		return
	}

	// Open the downloaded image based on the OS
	err = openImage(filepath)
	if err != nil {
		fmt.Printf("Error opening image: %v\n", err)
		return
	}
}