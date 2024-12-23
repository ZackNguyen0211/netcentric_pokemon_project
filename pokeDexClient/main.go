package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"bufio"
)

// Define the structure of the Pokémon data (same as the server side)
type Pokemon struct {
	Name      string `json:"name"`
	Height    int    `json:"height"`
	Weight    int    `json:"weight"`
	Types     []struct {
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

// Function to fetch Pokémon data from the server and display it
func fetchPokemon(name string) {
	// Normalize the name
	name = strings.ToLower(name)

	// Create the request URL
	url := fmt.Sprintf("http://localhost:8080/pokemon?name=%s", name)

	// Send the GET request
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	// Handle if Pokémon data is not found or other errors
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %v\n", resp.Status)
		return
	}

	// Decode the response
	var pokemon Pokemon
	if err := json.NewDecoder(resp.Body).Decode(&pokemon); err != nil {
		fmt.Printf("Failed to decode response: %v\n", err)
		return
	}

	// Print the Pokémon data
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)

	// Print types
	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("- %s\n", t.Type.Name)
	}

	// Print abilities
	fmt.Println("Abilities:")
	for _, a := range pokemon.Abilities {
		fmt.Printf("- %s\n", a.Ability.Name)
	}

	// Print stats
	fmt.Println("Stats:")
	for _, s := range pokemon.Stats {
		fmt.Printf("- %s: %d\n", s.Stat.Name, s.BaseStat)
	}
}

func main() {
	// Create a reader to read user input
	reader := bufio.NewReader(os.Stdin)

	// Loop for repeated input
	for {
		// Prompt the user to enter a Pokémon name
		fmt.Print("Enter the name of the Pokémon (or 'exit' to quit): ")
		input, _ := reader.ReadString('\n')
		name := strings.TrimSpace(input)

		// Exit if the user types 'exit'
		if strings.ToLower(name) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		// Fetch and display Pokémon data
		fetchPokemon(name)
	}
}
