from typing import List, Tuple
from sklearn.metrics import accuracy_score, precision_score, recall_score, roc_auc_score

def calculate_accuracy(true_outcomes: list[int], predicted_probabilities: list[int]) -> float:
  predictions = [1 if probability >= 0.5 else 0 for probability in predicted_probabilities]
  correct = sum(1 for true, pred in zip(true_outcomes, predictions) if true == pred)
  return correct / len(true_outcomes)

def sophisticated_accuracy(true_outcomes: list[int], predicted_probabilities: list[int]) -> dict[str, float]:
  return {
    'accuracy': accuracy_score(true_outcomes, [1 if p > 0.5 else 0 for p in predicted_probabilities]),
    'precision': precision_score(true_outcomes, [1 if p > 0.5 else 0 for p in predicted_probabilities]),
    'recall': recall_score(true_outcomes, [1 if p > 0.5 else 0 for p in predicted_probabilities]),
    'roc_auc': roc_auc_score(true_outcomes, predicted_probabilities),
  }
  
