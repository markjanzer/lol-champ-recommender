import numpy as np
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
from sklearn.neural_network import MLPClassifier
from pandas.core.frame import DataFrame
from typing import Tuple, Dict
from validations import sophisticated_accuracy

class NeuralNetChampionPredictor:
    def __init__(self, champions: DataFrame) -> None:
        self.num_champions = len(champions)
        self.champions = champions
        self.champion_to_index = {
            champ_id: idx for idx, champ_id in enumerate(champions['api_id'])
        }
        self.scaler = StandardScaler()
        self.threshold = 0.5  # Will be adjusted based on data
        
        # Initialize neural network with similar architecture to our PyTorch version
        self.model = MLPClassifier(
            hidden_layer_sizes=(256, 128),        # Two hidden layers
            activation='relu',                  # ReLU activation
            solver='adam',                      # Adam optimizer
            alpha=0.001,                        # Increased L2 regularization
            batch_size=32,                      # Mini-batch size
            learning_rate='adaptive',           # Adaptive learning rate
            max_iter=50,                        # Maximum epochs
            early_stopping=True,                # Enable early stopping
            validation_fraction=0.1,            # Validation set size
            n_iter_no_change=5,                 # Patience for early stopping
            random_state=42,
            verbose=True
        )
        
    def match_to_vector(self, match: DataFrame) -> np.array:
        vector = np.zeros(self.num_champions * 2)  # blue picks + red picks
        
        for i in range(5):
            blue_id = match[f"blue_{i+1}_champion_id"]
            red_id = match[f"red_{i+1}_champion_id"]
            
            blue_idx = self.champion_to_index[blue_id]
            red_idx = self.champion_to_index[red_id]
            
            vector[blue_idx] = 1
            vector[red_idx + self.num_champions] = 1
                
        return vector

    def prepare_data(self, matches: DataFrame) -> Tuple[np.array, np.array]:
        """Prepare match data for training and testing."""
        X = np.array([self.match_to_vector(matches.iloc[i]) for i in range(len(matches))])
        y = matches['winning_team'].eq('blue').to_numpy()
        return X, y

    def train_and_evaluate(self, matches: DataFrame, test_size: float = 0.1) -> Dict:
        """Train model and evaluate using sophisticated accuracy metrics."""
        # Prepare data
        X, y = self.prepare_data(matches)
        
        # Split data
        X_train, X_test, y_train, y_test = train_test_split(
            X, y, test_size=test_size, random_state=42
        )
        
        # Scale features
        X_train = self.scaler.fit_transform(X_train)
        X_test = self.scaler.transform(X_test)
        
        # Calculate class weights based on actual distribution
        blue_win_rate = np.mean(y_train)
        
        # Print diagnostics
        print("\nClass Distribution:")
        print(f"Blue team wins: {blue_win_rate:.4f}")
        print(f"Red team wins: {1 - blue_win_rate:.4f}")
        
        # Train model with class weights
        self.model.fit(X_train, y_train)
        
        # Get probability predictions
        y_pred_proba = self.model.predict_proba(X_test)[:, 1]
        
        # Set threshold based on actual win rate
        self.threshold = blue_win_rate
        
        # Calculate metrics using adjusted threshold
        metrics = sophisticated_accuracy(y_test, y_pred_proba)
        
        # Print prediction distribution
        predictions = y_pred_proba > self.threshold
        print("\nPrediction Distribution:")
        print(f"Predicted blue wins: {np.mean(predictions):.4f}")
        print(f"Predicted red wins: {1 - np.mean(predictions):.4f}")
        
        # Add convergence information
        metrics['n_iterations'] = self.model.n_iter_
        metrics['loss'] = self.model.loss_
        metrics['threshold'] = self.threshold
        
        return metrics
    
    def predict_winrate(self, match: DataFrame) -> float:
        """Predict win probability for given team composition."""
        vector = self.match_to_vector(match)
        vector_scaled = self.scaler.transform(vector.reshape(1, -1))
        win_prob = self.model.predict_proba(vector_scaled)[0, 1]
        # Return raw probability without thresholding for winrate predictions
        return win_prob
        
    def predict_winner(self, match: DataFrame) -> str:
        """Predict winning team using the calibrated threshold."""
        win_prob = self.predict_winrate(match)
        return "blue" if win_prob > self.threshold else "red"

if __name__ == "__main__":
    from utils.db_connector import get_all_champions, get_all_matches
    
    matches = get_all_matches()
    champions = get_all_champions()
    
    # Initialize predictor
    predictor = NeuralNetChampionPredictor(champions)
    
    # Train and evaluate
    results = predictor.train_and_evaluate(matches)
    
    # Print results
    print("\nNeural Network Model Performance:")
    for metric, value in results.items():
        print(f"{metric}: {value:.4f}")