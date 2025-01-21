# Lol Champ Recommender
This takes the current ally and enemy champions and recommends champions to play.

## Organization
`/go`  populates the database with matches, and generates the champion_stats data that is used by the statistical model.

`/python` validates the the model accuracy by looking at full matches and their results and using the models to predict the winner. This has several currently defunct machine learning models.

`/next` is the website.

## Usage
To run go commands, navigate to the go directory and run
```bash
go run cmd/<command>/main.go
```

To run a python model, navigate to the python directory and run
```bash
pipenv run python3 -m lolrecommender.models.<model_name>
```

## Updating the website's data
```bash
cd go
go run cmd/create_champion_stats/main.go # Run this for however long to seed data
go run cmd/write_json_to_next/main.go
cd ../next
vercel
```

## Go commands
These are largely internal notes for me to keep track of what I'm doing.

**create_champions**
This reads from a online file and creates the champions with a name and a api id.

**api_crawler**
This crawls the API for new matches and saves them.

It keeps track of which players have been crawled and how recently. Will crawl them again after 21 days.
It starts searching for new players to crawl from the existing saved matches. If there is none then it uses a seed player from the seed_accounts.json file.
When it finds a player it iterates over its past matches and saves them. They look like this:
```
type CreateMatchParams struct {
	MatchID         string
	GameStart       pgtype.Timestamp
	GameVersion     string
	WinningTeam     string
	Blue1ChampionID int32
	Blue2ChampionID int32
	Blue3ChampionID int32
	Blue4ChampionID int32
	Blue5ChampionID int32
	Red1ChampionID  int32
	Red2ChampionID  int32
	Red3ChampionID  int32
	Red4ChampionID  int32
	Red5ChampionID  int32
}
```
Right now only ranked matches are saved.


**create_champion_stats**
This reads all of the existing matches and creates a new ChampionStats object.
The jsonb of this object looks like this:
```
{
  01: {
    Matchups: {
      02: {
        Wins: 10,
        Games: 20
      },
      03: {...}
    },
    Synergies: {
      02: {
        Wins: 10,
        Games: 20
      },
      03: {...}
    }
  }
}
```


**champ_recommender**
This reads from the last champion stats object.
It looks at all of the selected champions (with and against) and then (right now) it average the winrates for all synergies and matchups to determine the winrate for the given champion with this composition.
For each champion it returns the overall averaged winrate, and then the synergies and matchups with their winrates.

**reset_db**
This drops all of the tables and creates new ones (except for champions)

**write_json_to_next**
This writes the champions and champion_stats to the nextjs data folder to be used by the website.
```

