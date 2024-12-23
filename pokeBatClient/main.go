package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type BattleRequest struct {
	Player1Pokemon []string `json:"player1_pokemon"`
	Player2Pokemon []string `json:"player2_pokemon"`
}

type ActionRequest struct {
	PlayerID string `json:"player_id"`
	Action   string `json:"action"`
}

type BattleState struct {
	Player1 struct {
		Name    string `json:"name"`
		Pokemon []struct {
			Name string `json:"name"`
			HP   int    `json:"hp"`
		} `json:"pokemon"`
	} `json:"player1"`
	Player2 struct {
		Name    string `json:"name"`
		Pokemon []struct {
			Name string `json:"name"`
			HP   int    `json:"hp"`
		} `json:"pokemon"`
	} `json:"player2"`
	Turn int `json:"turn"`
}

const serverURL = "http://localhost:8080"
var playerID string

// Helper function to send POST requests
func postRequest(endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(serverURL+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

// Fetch updated battle state from the server
func fetchBattleState() (BattleState, error) {
	resp, err := http.Get(serverURL + "/battle")
	if err != nil {
		return BattleState{}, fmt.Errorf("failed to fetch battle state: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return BattleState{}, fmt.Errorf("no active battle found")
	}

	var battleState BattleState
	if err := json.NewDecoder(resp.Body).Decode(&battleState); err != nil {
		return BattleState{}, fmt.Errorf("failed to decode battle state: %v", err)
	}
	return battleState, nil
}

// Function to initiate the battle if not already started
func startBattle() {
	fmt.Println("No active battle found. Would you like to start a new battle? (yes/no):")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "yes" {
		fmt.Println("Exiting the game.")
		os.Exit(0)
	}

	// Example Pokemon setup for a new battle
	battleRequest := BattleRequest{
		Player1Pokemon: []string{"Pikachu", "Charmander", "Bulbasaur"},
		Player2Pokemon: []string{"Squirtle", "Jigglypuff", "Meowth"},
	}
	_, err := postRequest("/start_battle", battleRequest)
	if err != nil {
		log.Fatalf("Failed to start a new battle: %v", err)
	}
	fmt.Println("Battle started successfully!")
}

func takeAction(battleState *BattleState) {
	for {
		// Fetch the updated battle state
		state, err := fetchBattleState()
		if err != nil {
			log.Printf("Error fetching battle state: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		*battleState = state

		fmt.Printf("Current turn: %d\n", battleState.Turn)
		fmt.Printf("Calculated turn odd: %t\n", battleState.Turn%2 == 1)
		fmt.Printf("Calculated turn even: %t\n", battleState.Turn%2 == 0)

		// Check the current player's turn
		if (battleState.Turn%2 == 1 )  {
			playerID = "player1"
			fmt.Printf("It's %s's turn\n", playerID)
			fmt.Println("Choose an action (attack/defend):")
			var action string
			fmt.Scanln(&action)

			// Send action request
			actionRequest := ActionRequest{
				PlayerID: playerID,
				Action:   strings.ToLower(action),
			}
			response, err := postRequest("/action", actionRequest)
			if err != nil {
				log.Printf("Failed to send action: %v", err)
				continue
			}

			// Decode updated battle state after action
			if err := json.Unmarshal(response, battleState); err != nil {
				log.Printf("Failed to decode battle state: %v", err)
				continue
			}
			break
		} else if (battleState.Turn%2 == 0) {
			playerID = "player2"
			fmt.Printf("It's %s's turn\n", playerID)
			fmt.Println("Choose an action (attack/defend):")
			var action string
			fmt.Scanln(&action)

			// Send action request
			actionRequest := ActionRequest{
				PlayerID: playerID,
				Action:   strings.ToLower(action),
			}
			response, err := postRequest("/action", actionRequest)
			if err != nil {
				log.Printf("Failed to send action: %v", err)
				continue
			}

			// Decode updated battle state after action
			if err := json.Unmarshal(response, battleState); err != nil {
				log.Printf("Failed to decode battle state: %v", err)
				continue
			}
			break
		} else {
			fmt.Println("Waiting for the other player to make a move...")
			time.Sleep(2 * time.Second) 
		}
	}
}

func main() {
	// Determine player ID
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run main.go [player1|player2]")
	}
	playerID = strings.ToLower(os.Args[1])
	if playerID != "player1" && playerID != "player2" {
		log.Fatalf("Invalid player ID. Use 'player1' or 'player2'")
	}

	// Fetch the initial battle state
	battleState, err := fetchBattleState()
	if err != nil {
		fmt.Println(err.Error())
		startBattle() // Start a new battle if none exists
		battleState, err = fetchBattleState()
		if err != nil {
			log.Fatalf("Failed to fetch battle state after starting a new battle: %v", err)
		}
	}

	// Game loop
	for {
		fmt.Printf("\nBattle State:\n")
		fmt.Printf("Player 1: %s\n", battleState.Player1.Name)
		for _, p := range battleState.Player1.Pokemon {
			fmt.Printf("- %s (HP: %d)\n", p.Name, p.HP)
		}
		fmt.Printf("Player 2: %s\n", battleState.Player2.Name)
		for _, p := range battleState.Player2.Pokemon {
			fmt.Printf("- %s (HP: %d)\n", p.Name, p.HP)
		}

		// Check for winner
		allFainted1 := true
		allFainted2 := true
		for _, p := range battleState.Player1.Pokemon {
			if p.HP > 0 {
				allFainted1 = false
				break
			}
		}
		for _, p := range battleState.Player2.Pokemon {
			if p.HP > 0 {
				allFainted2 = false
				break
			}
		}

		if allFainted1 {
			fmt.Println("Player 2 wins!")
			break
		} else if allFainted2 {
			fmt.Println("Player 1 wins!")
			break
		}

		takeAction(&battleState)
	}
}
