package champions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lol-champ-recommender/db"
	"net/http"
	"strconv"

	"lol-champ-recommender/internal/version"
)

type ChampionData struct {
	Data map[string]Champion `json:"data"`
}

type Champion struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// Turn a version like 14.18.618.2357 into 14.18.1
func toMajorMinorOne(version version.GameVersion) string {
	return fmt.Sprintf("%d.%d.1", version.Major, version.Minor)
}

func championsURL(version string) string {
	return fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion.json", version)
}

func upsertChampionsFromVersion(ctx context.Context, queries *db.Queries, version string) error {
	championsURL := championsURL(version)

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

func latestVersion(ctx context.Context, queries *db.Queries) (version.GameVersion, error) {
	versions, err := queries.GameVersions(ctx)
	if err != nil {
		return version.GameVersion{}, err
	}

	return version.GetLatest(versions)
}

func UpsertChampions(ctx context.Context, queries *db.Queries) error {
	version, err := latestVersion(ctx, queries)
	if err != nil {
		return err
	}

	shortenedVersion := toMajorMinorOne(version)
	fmt.Println("Upserting champions from version", shortenedVersion)

	err = upsertChampionsFromVersion(ctx, queries, shortenedVersion)
	return err
}
