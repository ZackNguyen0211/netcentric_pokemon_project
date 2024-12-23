package gameplay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Move struct {
	Name       string
	Damage     int
	Special    bool
}

type Pokemon struct {
	Name          string `json:"name"`
	Height        int    `json:"height"`
	Weight        int    `json:"weight"`
	HP            int
	Attack        int
	Special       int
	Speed         int
	Stats         []struct {
		Stat struct {
			Name string `json:"name"`
		} `json:"stat"`
		BaseStat int `json:"base_stat"`
	} `json:"stats"`
	Types         []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	Abilities     []struct {
		Ability struct {
			Name string `json:"name"`
		} `json:"ability"`
	} `json:"abilities"`
	DefenseBoost  int
}

type Player struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Pokemon             []Pokemon  `json:"pokemon"`
	CurrentPokemonIndex int        `json:"current_pokemon_index"`
}

type Battle struct {
	Player1 Player `json:"player1"`
	Player2 Player `json:"player2"`
	Turn     int    `json:"turn"`
}

func ReadPokemonData(number string) (Pokemon, error) {
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

	var rawPokemon Pokemon
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("Error: Failed to read data from file %s: %v", filename, err)
		return Pokemon{}, fmt.Errorf("failed to read data from file %s: %v", filename, err)
	}

	if err := json.Unmarshal(data, &rawPokemon); err != nil {
		log.Printf("Error: Failed to unmarshal JSON for file %s: %v", filename, err)
		return Pokemon{}, fmt.Errorf("failed to unmarshal JSON for file %s: %v", filename, err)
	}

	// Map stats to individual fields
	pokemon := Pokemon{
		Name:      rawPokemon.Name,
		Height:    rawPokemon.Height,
		Weight:    rawPokemon.Weight,
		Types:     rawPokemon.Types,
		Abilities: rawPokemon.Abilities,
	}
	for _, stat := range rawPokemon.Stats {
		switch stat.Stat.Name {
		case "hp":
			pokemon.HP = stat.BaseStat
		case "attack":
			pokemon.Attack = stat.BaseStat
		case "special-attack":
			pokemon.Special = stat.BaseStat
		case "speed":
			pokemon.Speed = stat.BaseStat
		}
	}

	log.Printf("Successfully read data for Pokémon %s (HP: %d, Attack: %d, Special: %d, Speed: %d)", pokemon.Name, pokemon.HP, pokemon.Attack, pokemon.Special, pokemon.Speed)
	return pokemon, nil
}

func hasRemainingPokemon(player *Player) bool {
	for _, pkmn := range player.Pokemon {
		if pkmn.HP > 0 {
			return true
		}
	}
	return false
}

func (pokemon *Pokemon) Defense() int {
	return pokemon.Weight / 10
}

func ExecuteAttack(battle *Battle, playerID string) {
	var attacker *Pokemon
	var defender *Pokemon
	var currentPlayer *Player
	var opposingPlayer *Player

	if playerID == "player1" {
		currentPlayer = &battle.Player1
		opposingPlayer = &battle.Player2
	} else {
		currentPlayer = &battle.Player2
		opposingPlayer = &battle.Player1
	}

	// Get the current Pokémon for both players
	attacker = &currentPlayer.Pokemon[currentPlayer.CurrentPokemonIndex]
	defender = &opposingPlayer.Pokemon[opposingPlayer.CurrentPokemonIndex]

	move := Move{Name: "Tackle", Damage: 40, Special: false}

	var damage int
	if move.Special {
		// Special move damage
		damage = attacker.Special + move.Damage - defender.Defense()
	} else {
		// Regular move damage
		damage = attacker.Attack + move.Damage - defender.Defense()
	}

	// Apply defense boost if defender has it
	if defender.DefenseBoost > 0 {
		damage -= defender.DefenseBoost
		if damage < 1 {
			damage = 1
		}
		defender.DefenseBoost = 0
	}

	// Apply damage to the defender's HP
	defender.HP -= damage
	if defender.HP < 0 {
		defender.HP = 0
	}

	log.Printf("%s attacked %s with %s, causing %d damage!", attacker.Name, defender.Name, move.Name, damage)

	// If the defender's Pokémon fainted, move to the next one
	if defender.HP <= 0 {
		opposingPlayer.CurrentPokemonIndex++
	}

	// Check if the opposing player has any remaining Pokémon
	if !hasRemainingPokemon(opposingPlayer) {
		log.Printf("%s has no remaining Pokémon! %s wins!", opposingPlayer.Name, currentPlayer.Name)
		// Handle game win condition here
		return
	}
}

// Function to execute the defend action (new)
func ExecuteDefend(battle *Battle, playerID string) {
	var defender *Pokemon
	var currentPlayer *Player

	// Determine the current player and opposing player based on playerID
	if playerID == "player1" {
		currentPlayer = &battle.Player1
	} else {
		currentPlayer = &battle.Player2
	}

	// Get the current Pokémon for both players
	defender = &currentPlayer.Pokemon[currentPlayer.CurrentPokemonIndex]

	// Increase the defender's defense boost for the next attack
	defender.DefenseBoost += 100

	log.Printf("%s chose to defend! %s's defense will be boosted on the next attack.", defender.Name, defender.Name)
}