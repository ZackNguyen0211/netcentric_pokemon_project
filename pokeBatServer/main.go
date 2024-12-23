package main

import (
	"encoding/json"
	"log"
	"net/http"
	"netcentric/gameplay"
	"strings"
	"netcentric/utils"
)

var battle *gameplay.Battle

// Handle the battle requests (starting a battle)
func handleBattleRequest(w http.ResponseWriter, r *http.Request) {
	// Ensure method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the incoming battle request
	var battleRequest struct {
		Player1Pokemon []string `json:"player1_pokemon"`
		Player2Pokemon []string `json:"player2_pokemon"`
	}
	if err := json.NewDecoder(r.Body).Decode(&battleRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Initialize Players
	player1 := gameplay.Player{ID: "player1", Name: "Player 1", Pokemon: make([]gameplay.Pokemon, 3)}
	player2 := gameplay.Player{ID: "player2", Name: "Player 2", Pokemon: make([]gameplay.Pokemon, 3)}

	// Fetch Pokémon data based on player selection
	for i, pkmnName := range battleRequest.Player1Pokemon {
		number, _ := utils.PokeMap[strings.Title(strings.ToLower(pkmnName))]
		pokemon, _ := gameplay.ReadPokemonData(number)
		player1.Pokemon[i] = pokemon
	}

	for i, pkmnName := range battleRequest.Player2Pokemon {
		number, _ := utils.PokeMap[strings.Title(strings.ToLower(pkmnName))]
		pokemon, _ := gameplay.ReadPokemonData(number) 
		player2.Pokemon[i] = pokemon
	}

	// Start Battle
	battle = &gameplay.Battle{
		Player1: player1,
		Player2: player2,
		Turn:    1,
	}

	// Respond with the initial battle state
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(battle)
}

// Handle fetching the current battle state GET method
func handleBattleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if battle == nil {
		http.Error(w, "No active battle", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(battle)
}

// Handle actions for each turn
func handleAction(w http.ResponseWriter, r *http.Request) {
	// Ensure method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode action request
	var actionRequest struct {
		PlayerID string `json:"player_id"`
		Action   string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&actionRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if the current turn matches the requesting player
	isPlayer1Turn := battle.Turn%2 == 1
	if (actionRequest.PlayerID == "player1" && !isPlayer1Turn) ||
		(actionRequest.PlayerID == "player2" && isPlayer1Turn) {
		http.Error(w, "Not your turn", http.StatusBadRequest)
		return
	}

	// Execute turn logic
	switch strings.ToLower(actionRequest.Action) {
	case "attack":
		gameplay.ExecuteAttack(battle, actionRequest.PlayerID)
	case "defend":
		gameplay.ExecuteDefend(battle, actionRequest.PlayerID)
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	// Increment turn
	battle.Turn++

	// Respond with updated battle state
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(battle)
}

// Main function to start the server
func main() {

	http.HandleFunc("/start_battle", handleBattleRequest)
	http.HandleFunc("/battle", handleBattleState)
	http.HandleFunc("/action", handleAction)

	// Start the server
	log.Println("Starting battle server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
