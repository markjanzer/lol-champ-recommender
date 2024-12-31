from winrates import average_prediction, weighted_prediction
from validations import calculate_accuracy, sophisticated_accuracy
from utils.db_connector import get_champion_stats, get_matches_above_id
import pandas as pd

def main():
  data = get_champion_stats()
  champion_stats = data.data[0]
  last_match_id = data.last_match_id[0]
  matches = get_matches_above_id(last_match_id)
  outcomes = [1 if match["winning_team"] == "blue" else 0 for _, match in matches.iterrows()]

  average_predictions = [average_prediction(pd.DataFrame([match]), champion_stats) for _, match in matches.iterrows()]
  weighted_predictions = [weighted_prediction(pd.DataFrame([match]), champion_stats) for _, match in matches.iterrows()]

  print("Average Accuracy: ", sophisticated_accuracy(outcomes, average_predictions))
  print("Weighted Accuracy: ", sophisticated_accuracy(outcomes, weighted_predictions))

if __name__ == "__main__":
  main()