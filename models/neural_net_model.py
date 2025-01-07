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
        self.scaler = StandardScaler()
        
        # Initialize neural network with similar architecture to our PyTorch version
        self.model = MLPClassifier(
            hidden_layer_sizes=(256, 128, 64),  # Three hidden layers
            activation='relu',                  # ReLU activation
            solver='adam',                      # Adam optimizer
            alpha=0.0001,                       # L2 regularization
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
        """Convert match data into feature vector."""
        vector = np.zeros(self.num_champions * 2)  # blue picks + red picks
        
        blue_keys = [f"blue_{i+1}_champion_id" for i in range(5)]
        red_keys = [f"red_{i+1}_champion_id" for i in range(5)]

        for key in blue_keys:
            pick_id = match[key]
            pick_index = self.champions[self.champions['api_id'] == pick_id].index[0]
            if 0 <= pick_index < self.num_champions:
                vector[pick_index] = 1
                
        for key in red_keys:
            pick_id = match[key]
            pick_index = self.champions[self.champions['api_id'] == pick_id].index[0]
            if 0 <= pick_index < self.num_champions:
                vector[pick_index + self.num_champions] = 1
                
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
        
        # Train model
        self.model.fit(X_train, y_train)
        
        # Get probability predictions
        y_pred_proba = self.model.predict_proba(X_test)[:, 1]
        
        # Calculate metrics
        metrics = sophisticated_accuracy(y_test, y_pred_proba)
        
        # Add convergence information
        metrics['n_iterations'] = self.model.n_iter_
        metrics['loss'] = self.model.loss_
        
        return metrics
    
    def predict_winrate(self, match: DataFrame) -> float:
        """Predict win probability for given team composition."""
        vector = self.match_to_vector(match)
        vector_scaled = self.scaler.transform(vector.reshape(1, -1))
        win_prob = self.model.predict_proba(vector_scaled)[0, 1]
        return win_prob

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
        if isinstance(value, float):
            print(f"{metric}: {value:.4f}")
        else:
            print(f"{metric}: {value}")