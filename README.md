# Lol Champ Recommender

Trying a variety of approaches to see which is the most effective
at recommending champions for a given match. It should work with partial drafts, and it should also be able to 

## What it does
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
I'm not sure what types of matches are saved.


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
Still not 100% sure how much data is returned here, do we offer matchup specific data? I think so but I'm not sure.

**reset_db**
This drops all of the tables and creates new ones (except for champions)
