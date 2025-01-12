package recommender

import (
	"context"
	"fmt"
	"lol-champ-recommender/db"
	"strings"
)

func FormatAnswer(ctx context.Context, queries *db.Queries, champSelect ChampSelect, results []ChampionPerformance) error {
	champsToIDs, err := mapChampionsToIds(ctx, queries)
	if err != nil {
		return fmt.Errorf("error mapping champions to IDs: %v", err)
	}

	bannedChamps := []string{}
	for _, ban := range champSelect.Bans {
		bannedChamps = append(bannedChamps, IDToName(champsToIDs, ban))
	}
	bannedChampsString := strings.Join(bannedChamps, ", ")
	fmt.Println("Bans:", bannedChampsString)

	allyChamps := []string{}
	for _, ally := range champSelect.Allies {
		allyChamps = append(allyChamps, IDToName(champsToIDs, ally))
	}
	allyChampsString := strings.Join(allyChamps, ", ")
	fmt.Println("Allies:", allyChampsString)

	enemyChamps := []string{}
	for _, enemy := range champSelect.Enemies {
		enemyChamps = append(enemyChamps, IDToName(champsToIDs, enemy))
	}
	enemyChampsString := strings.Join(enemyChamps, ", ")
	fmt.Println("Enemies:", enemyChampsString)

	fmt.Println("Recommended:")
	for _, result := range results {
		printChampionPerformance(champsToIDs, result)
	}

	return nil
}

func mapChampionsToIds(ctx context.Context, queries *db.Queries) (map[string]int32, error) {
	champions, err := queries.AllChampions(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting all champions: %v", err)
	}

	result := make(map[string]int32)
	for _, champ := range champions {
		result[champ.Name] = champ.ApiID
	}

	return result, nil
}

func printChampionPerformance(champsToIDs map[string]int32, champion ChampionPerformance) {
	championName := IDToName(champsToIDs, champion.ChampionID)
	winPercentage := probabilityAsPercentage(champion.WinProbability)
	matchupsString := printChampionInteractions(champsToIDs, champion.Matchups)
	synergiesString := printChampionInteractions(champsToIDs, champion.Synergies)

	fmt.Printf("%s: %s — MATCHUPS [ %s ] — SYNERGIES [ %s ]\n",
		championName,
		winPercentage,
		matchupsString,
		synergiesString)
}

func probabilityAsPercentage(probability float64) string {
	return fmt.Sprintf("%.2f%%", probability*100)
}

func printChampionInteractions(champsToIDs map[string]int32, interactions []ChampionInteraction) string {
	var result []string
	for _, interaction := range interactions {
		championName := IDToName(champsToIDs, interaction.ChampionID)
		result = append(result, fmt.Sprintf("%s: %d/%d", championName, interaction.Wins, interaction.Games))
	}
	return strings.Join(result, ", ")
}

func IDToName(champions map[string]int32, id int32) string {
	for name, champID := range champions {
		if champID == id {
			return name
		}
	}
	return "Unknown"
}
