package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"lol-champ-recommender/db"
	"lol-champ-recommender/internal/database"
	"net/http"
	"strconv"
	"strings"
)

type ChampionData struct {
	Data map[string]Champion `json:"data"`
}

type Champion struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// Turn a version like 14.18.618.2357 into 14.18.1
func simplifyVersion(version string) string {
	versionNumbers := strings.Split(version, ".")
	simplifiedVersion := fmt.Sprintf("%s.%s.1", versionNumbers[0], versionNumbers[1])
	return simplifiedVersion
}

func championsURL(version string) string {
	return fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion.json", version)
}

func UpdateChampionsFromVersion(ctx context.Context, queries *db.Queries, version string) error {
	simplifiedVersion := simplifyVersion(version)
	championsURL := championsURL(simplifiedVersion)

	resp, err := http.Get(championsURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var championData ChampionData
	err = json.Unmarshal(body, &championData)
	if err != nil {
		return err
	}

	for _, champion := range championData.Data {
		apiID, err := strconv.Atoi(champion.Key)
		if err != nil {
			return err
		}

		err = queries.UpsertChampion(ctx, db.UpsertChampionParams{
			Name:  champion.Name,
			ApiID: int32(apiID),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateChampions(ctx context.Context, queries *db.Queries) error {
	lastMatch, err := queries.LastMatch(ctx)
	if err != nil {
		return err
	}

	return UpdateChampionsFromVersion(ctx, queries, lastMatch.GameVersion)
}

// Search the matches for the latest version, update the champions from that version
func main() {
	ctx := context.Background()

	dbConn, err := database.Initialize(ctx)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer dbConn.Close(ctx)

	err = UpdateChampions(ctx, dbConn.Queries)
	if err != nil {
		log.Fatalf("Error updating champions: %v", err)
	}
}
