import pandas as pd
import itertools
from typing import TypedDict, Tuple


def get_teams(match):
  blue_team = [
    match.iloc[0]['blue_1_champion_id'] ,
    match.iloc[0]['blue_2_champion_id'],
    match.iloc[0]['blue_3_champion_id'],
    match.iloc[0]['blue_4_champion_id'],
    match.iloc[0]['blue_5_champion_id']
  ]

  red_team = [
    match.iloc[0]['red_1_champion_id'],
    match.iloc[0]['red_2_champion_id'],
    match.iloc[0]['red_3_champion_id'],
    match.iloc[0]['red_4_champion_id'],
    match.iloc[0]['red_5_champion_id']
  ]

  return blue_team, red_team

class WinStats(TypedDict):
  wins: int
  games: int

def get_winrate(stats: WinStats) -> float:
  if stats['games'] == 0:
    return 0.5
  return smooth_winrate(stats['wins'], stats['games'], 5, 10)

def smooth_winrate(wins: int, games: int, prior_wins: int, prior_games: int) -> float:
  return (wins + prior_wins) / (games + prior_games)

def get_all_combinations(numbers):
  return list(itertools.combinations(numbers, 2))

def predict_win_with_average(blue_team_synergy: float, red_team_synergy: float, blue_team_matchup: float) -> float:
  red_factor = 1 - red_team_synergy
  return (blue_team_synergy + red_factor + blue_team_matchup) / 3

def predict_win_with_weighted_average(blue_team_synergy: float, red_team_synergy: float, blue_team_matchup: float) -> float:
  red_factor = 1 - red_team_synergy
  return (0.25 * blue_team_synergy + 0.25 * red_factor + 0.5 * blue_team_matchup)

def match_stats(match: pd.DataFrame, champion_stats: dict) -> Tuple[float, float, float]:
  blue_team, red_team = get_teams(match)

  blue_team_combinations = get_all_combinations(blue_team)
  red_team_combinations = get_all_combinations(red_team)

  blue_team_synergies = []
  red_team_synergies = []

  for combination in blue_team_combinations: 
    try:
      blue_team_synergies.append(champion_stats[str(combination[0])]['synergies'][str(combination[1])])
    except KeyError:
      print(f"KeyError for blue team synergies {combination}, match: {match}")
      raise KeyError


  for combination in red_team_combinations: 
    try:
      red_team_synergies.append(champion_stats[str(combination[0])]['synergies'][str(combination[1])])
    except KeyError:
      print(f"KeyError for red team synergies {combination}, match: {match}")
      raise KeyError
  blue_team_winrates = [get_winrate(synergy) for synergy in blue_team_synergies]
  red_team_winrates = [get_winrate(synergy) for synergy in red_team_synergies]

  blue_team_synergy = sum(blue_team_winrates) / len(blue_team_winrates)
  red_team_synergy = sum(red_team_winrates) / len(red_team_winrates)

  # This always has blue team first
  opposing_pairings = []
  for blue_team_champion in blue_team:
    for red_team_champion in red_team:
      opposing_pairings.append((blue_team_champion, red_team_champion))
  
  matchups_from_blue_team = []
  for matchup in opposing_pairings:
    try:
      matchups_from_blue_team.append(champion_stats[str(matchup[0])]['matchups'][str(matchup[1])])
    except KeyError:
      print(f"KeyError for matchup {matchup}, match: {match}")
      raise KeyError

  blue_team_winrates = [get_winrate(matchup) for matchup in matchups_from_blue_team]
  blue_team_matchup = sum(blue_team_winrates) / len(blue_team_winrates)

  return blue_team_synergy, red_team_synergy, blue_team_matchup

def average_prediction(match: pd.DataFrame, matchup_stats: dict) -> float:
  blue_team_synergy, red_team_synergy, blue_team_matchup = match_stats(match, matchup_stats)
  return predict_win_with_average(blue_team_synergy, red_team_synergy, blue_team_matchup)

def weighted_prediction(match: pd.DataFrame, matchup_stats: dict) -> float:
  blue_team_synergy, red_team_synergy, blue_team_matchup = match_stats(match, matchup_stats)
  return predict_win_with_weighted_average(blue_team_synergy, red_team_synergy, blue_team_matchup)


