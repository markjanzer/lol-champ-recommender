import numpy as np
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
from sklearn.linear_model import LogisticRegression
from validations import sophisticated_accuracy
from utils.db_connector import get_all_champions, get_all_matches
from pandas.core.frame import DataFrame
from typing import Tuple

class ChampionPredictor:
    def __init__(self, champions: DataFrame) -> None:
        self.num_champions = len(champions)
        self.champions = champions
        self.scaler = StandardScaler()
        self.model = LogisticRegression(max_iter=1000)
        
    def match_to_vector(self, match: DataFrame) -> np.array:
        """Convert match data into feature vector."""
        vector = np.zeros(self.num_champions * 2)  # blue picks + red picks
        
        blue_keys = [f"blue_{i+1}_champion_id" for i in range(5)]
        red_keys = [f"red_{i+1}_champion_id" for i in range(5)]

        for key in blue_keys:
            pick_id = match[key]
            pick_index = self.champions[self.champions['api_id'] == pick_id].index[0]
            if 0 <= pick_index < self.num_champions:
                vector[pick_index] = 1
                
        # Set red picks (second N positions)
        for key in red_keys:
            pick_id = match[key]
            pick_index = self.champions[self.champions['api_id'] == pick_id].index[0]
            if 0 <= pick_index < self.num_champions:
                vector[pick_index + self.num_champions] = 1
                
        return vector

    def prepare_data(self, matches: DataFrame) -> Tuple[np.array, np.array]:
        """
        Prepare match data for training and testing.
        
        Args:
            matches: List of match dictionaries with 'blue_picks', 'red_picks', and 'blue_won'
            
        Returns:
            X: Feature matrix
            y: Binary labels (1 if blue team won, 0 if red team won)
        """
        # Convert matches to feature vectors
        X = np.array([self.match_to_vector(matches.iloc[i]) for i in range(len(matches))])
        
        # Create labels from match outcomes
        y = matches['winning_team'].eq('blue').to_numpy()
        
        return X, y

    def train_and_evaluate(self, matches: DataFrame, test_size: float = 0.1) -> dict:
        """
        Train model and evaluate using the sophisticated accuracy metrics.
        
        Args:
            matches: List of match dictionaries
            test_size: Fraction of data to use for testing
            
        Returns:
            dict: Dictionary of evaluation metrics
        """
        # Prepare data
        X, y = self.prepare_data(matches)
        
        # Split data
        X_train, X_test, y_train, y_test = train_test_split(
            X, y, test_size=test_size, random_state=42
        )
        
        # Scale features
        X_train = self.scaler.fit_transform(X_train)
        X_test = self.scaler.transform(X_test)
        
        # Train model
        self.model.fit(X_train, y_train)
        
        # Get probability predictions
        y_pred_proba = self.model.predict_proba(X_test)[:, 1]
        
        # Use sophisticated accuracy function
        metrics = sophisticated_accuracy(y_test, y_pred_proba)
        
        return metrics
    
    def predict_winrate(self, match: DataFrame) -> float:
        """
        Predict the win probability for a given team composition.
        
        Args:
            blue_picks: List of champion IDs for blue team
            red_picks: List of champion IDs for red team
            
        Returns:
            float: Probability of blue team winning
        """
        # Create match vector
        vector = self.match_to_vector(match)
        
        # Scale the vector
        vector_scaled = self.scaler.transform(vector.reshape(1, -1))
        
        # Get win probability
        win_prob = self.model.predict_proba(vector_scaled)[0, 1]
        
        return win_prob

# Example usage
if __name__ == "__main__":
    matches = get_all_matches()
    champions = get_all_champions()
    
    # Initialize predictor
    predictor = ChampionPredictor(champions)
    
    # Train and evaluate
    results = predictor.train_and_evaluate(matches)
    
    # Print results
    print("\nLogistic Regression Performance:")
    for metric, value in results.items():
        print(f"{metric}: {value:.4f}")