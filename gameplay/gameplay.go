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
	Name      string `json:"name"`
	Height    int    `json:"height"`
	Weight    int    `json:"weight"`
	HP        int
	Attack    int
	Special   int
	Speed     int
	Stats     []struct {
		Stat struct {
			Name string `json:"name"`
		} `json:"stat"`
		BaseStat int `json:"base_stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	Abilities []struct {
		Ability struct {
			Name string `json:"name"`
		} `json:"ability"`
	} `json:"abilities"`
}

type Player struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Pokemon             []Pokemon `json:"pokemon"`
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


// Function to get Pokémon's defense stat (simplified for now)
func (pokemon *Pokemon) Defense() int {
	// Simplified defense
	return pokemon.Weight / 10
}

// Function to check if a player has any remaining Pokémon
func hasRemainingPokemon(player *Player) bool {
	log.Printf("Checking remaining Pokémon for %s:", player.Name)
	for _, pkmn := range player.Pokemon {
		log.Printf("  %s HP: %d", pkmn.Name, pkmn.HP)
		if pkmn.HP > 0 {
			return true
		}
	}
	return false
}


// Function to handle player attacks
func attack(attacker *Pokemon, defender *Pokemon, move Move) {
	var damage int
	if move.Special {
		// Special attack
		damage = attacker.Special - defender.Defense()
	} else {
		// Normal attack
		damage = attacker.Attack - defender.Defense()
	}
	if damage < 1 {
		damage = 1
	}

	defender.HP -= damage
	if defender.HP < 0 {
		defender.HP = 0
	}
	log.Printf("%s attacked %s with %s, causing %d damage!", attacker.Name, defender.Name, move.Name, damage)
}

// Function to execute the turn and switch between players
func ExecuteTurn(battle *Battle) {
	// Check whose turn it is
	if battle.Turn%2 == 1 {
		// Player 1's turn
		log.Printf("Player 1's turn!")

		// Example of a player action (attack, move, etc.)
		// Assume we are just attacking with the first move
		move := Move{Name: "Tackle", Damage: 40, Special: false}
		attack(&battle.Player1.Pokemon[battle.Player1.CurrentPokemonIndex], &battle.Player2.Pokemon[battle.Player2.CurrentPokemonIndex], move)

		// Check if Player 2's Pokémon is fainted, if so, move to the next one
		if battle.Player2.Pokemon[battle.Player2.CurrentPokemonIndex].HP <= 0 {
			battle.Player2.CurrentPokemonIndex++
		}

	} else {
		// Player 2's turn
		log.Printf("Player 2's turn!")

		// Example of a player action (attack, move, etc.)
		// Assume we are just attacking with the first move
		move := Move{Name: "Tackle", Damage: 40, Special: false}
		attack(&battle.Player2.Pokemon[battle.Player2.CurrentPokemonIndex], &battle.Player1.Pokemon[battle.Player1.CurrentPokemonIndex], move)

		// Check if Player 1's Pokémon is fainted, if so, move to the next one
		if battle.Player1.Pokemon[battle.Player1.CurrentPokemonIndex].HP <= 0 {
			battle.Player1.CurrentPokemonIndex++
		}
	}

	// Check if Player 1 has any remaining Pokémon
	if !hasRemainingPokemon(&battle.Player1) {
		log.Printf("Player 1 has no remaining Pokémon! Player 2 wins!")
		// You can implement a win condition here and return if desired
		return
	}

	// Check if Player 2 has any remaining Pokémon
	if !hasRemainingPokemon(&battle.Player2) {
		log.Printf("Player 2 has no remaining Pokémon! Player 1 wins!")
		// You can implement a win condition here and return if desired
		return
	}

	// End of turn, increment turn counter
	battle.Turn++
}

