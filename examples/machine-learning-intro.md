# Introduction to Machine Learning

Explore the fascinating world of machine learning. This course covers fundamental concepts, algorithms, and practical applications using Python and scikit-learn.

## What is Machine Learning?

Machine Learning is a subset of artificial intelligence that enables computers to learn and make decisions from data without being explicitly programmed.

### Types of Machine Learning

#### Supervised Learning
- **Definition**: Learning from labeled training data
- **Goal**: Make predictions on new, unseen data
- **Examples**: Classification, Regression

#### Unsupervised Learning
- **Definition**: Finding patterns in unlabeled data
- **Goal**: Discover hidden structure in data
- **Examples**: Clustering, Dimensionality reduction

#### Reinforcement Learning
- **Definition**: Learning through interaction with environment
- **Goal**: Maximize cumulative reward
- **Examples**: Game playing, Robotics

## Setting Up Your Environment

### Installing Required Packages
```bash
# Create a virtual environment
python -m venv ml-env
source ml-env/bin/activate  # On Windows: ml-env\Scripts\activate

# Install packages
pip install numpy pandas matplotlib seaborn scikit-learn jupyter
```

### Alternative: Google Colab
- No installation required
- Free GPU access
- Pre-installed ML libraries
- Visit colab.research.google.com

## Data Preparation

### Loading Data
```python
import pandas as pd
import numpy as np

# Load sample datasets
from sklearn.datasets import load_iris, load_boston, make_classification

# Iris dataset (classification)
iris = load_iris()
X_iris, y_iris = iris.data, iris.target

# Boston housing dataset (regression)
boston = load_boston()
X_boston, y_boston = boston.data, boston.target

# Generate synthetic data
X_synth, y_synth = make_classification(n_samples=1000, n_features=20, n_classes=2, random_state=42)
```

### Data Preprocessing
```python
from sklearn.preprocessing import StandardScaler, LabelEncoder
from sklearn.model_selection import train_test_split

# Split data into training and testing sets
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

# Standardize features (mean=0, variance=1)
scaler = StandardScaler()
X_train_scaled = scaler.fit_transform(X_train)
X_test_scaled = scaler.transform(X_test)

# Encode categorical variables
encoder = LabelEncoder()
y_encoded = encoder.fit_transform(y_categorical)
```

## Supervised Learning: Classification

### K-Nearest Neighbors (KNN)
```python
from sklearn.neighbors import KNeighborsClassifier
from sklearn.metrics import accuracy_score, classification_report

# Create and train the model
knn = KNeighborsClassifier(n_neighbors=5)
knn.fit(X_train_scaled, y_train)

# Make predictions
y_pred = knn.predict(X_test_scaled)

# Evaluate the model
accuracy = accuracy_score(y_test, y_pred)
print(f"Accuracy: {accuracy:.3f}")

# Detailed classification report
print(classification_report(y_test, y_pred))
```

### Support Vector Machines (SVM)
```python
from sklearn.svm import SVC

# Linear SVM
svm_linear = SVC(kernel='linear', C=1.0)
svm_linear.fit(X_train_scaled, y_train)

# RBF kernel SVM
svm_rbf = SVC(kernel='rbf', C=1.0, gamma='scale')
svm_rbf.fit(X_train_scaled, y_train)

# Polynomial kernel SVM
svm_poly = SVC(kernel='poly', degree=3, C=1.0)
svm_poly.fit(X_train_scaled, y_train)
```

### Decision Trees
```python
from sklearn.tree import DecisionTreeClassifier
from sklearn.tree import plot_tree
import matplotlib.pyplot as plt

# Create and train decision tree
dt = DecisionTreeClassifier(max_depth=3, random_state=42)
dt.fit(X_train, y_train)

# Visualize the tree
plt.figure(figsize=(20,10))
plot_tree(dt, feature_names=iris.feature_names, class_names=iris.target_names, filled=True)
plt.show()

# Feature importance
feature_importance = pd.DataFrame({
    'feature': iris.feature_names,
    'importance': dt.feature_importances_
}).sort_values('importance', ascending=False)

print(feature_importance)
```

### Random Forest
```python
from sklearn.ensemble import RandomForestClassifier

# Create and train random forest
rf = RandomForestClassifier(n_estimators=100, max_depth=10, random_state=42)
rf.fit(X_train, y_train)

# Feature importance
feature_importance = pd.DataFrame({
    'feature': feature_names,
    'importance': rf.feature_importances_
}).sort_values('importance', ascending=False)

print(feature_importance.head(10))
```

### Gradient Boosting
```python
from sklearn.ensemble import GradientBoostingClassifier

# Create and train gradient boosting classifier
gb = GradientBoostingClassifier(n_estimators=100, learning_rate=0.1, max_depth=3, random_state=42)
gb.fit(X_train, y_train)

# Feature importance
feature_importance = pd.DataFrame({
    'feature': feature_names,
    'importance': gb.feature_importances_
}).sort_values('importance', ascending=False)

print(feature_importance.head(10))
```

## Supervised Learning: Regression

### Linear Regression
```python
from sklearn.linear_model import LinearRegression
from sklearn.metrics import mean_squared_error, r2_score

# Create and train linear regression model
lr = LinearRegression()
lr.fit(X_train_scaled, y_train)

# Make predictions
y_pred = lr.predict(X_test_scaled)

# Evaluate the model
mse = mean_squared_error(y_test, y_pred)
r2 = r2_score(y_test, y_pred)

print(f"Mean Squared Error: {mse:.2f}")
print(f"R² Score: {r2:.3f}")

# Coefficients
coefficients = pd.DataFrame({
    'feature': feature_names,
    'coefficient': lr.coef_
})
print(coefficients)
```

### Polynomial Regression
```python
from sklearn.preprocessing import PolynomialFeatures
from sklearn.pipeline import Pipeline

# Create polynomial features and linear regression pipeline
poly_reg = Pipeline([
    ('poly_features', PolynomialFeatures(degree=2)),
    ('linear_reg', LinearRegression())
])

poly_reg.fit(X_train_scaled, y_train)
y_pred = poly_reg.predict(X_test_scaled)

# Evaluate
mse = mean_squared_error(y_test, y_pred)
r2 = r2_score(y_test, y_pred)
print(f"Polynomial Regression - MSE: {mse:.2f}, R²: {r2:.3f}")
```

### Ridge and Lasso Regression
```python
from sklearn.linear_model import Ridge, Lasso

# Ridge regression (L2 regularization)
ridge = Ridge(alpha=1.0)
ridge.fit(X_train_scaled, y_train)

# Lasso regression (L1 regularization)
lasso = Lasso(alpha=0.1)
lasso.fit(X_train_scaled, y_train)

# Compare coefficients
ridge_coef = pd.DataFrame({'feature': feature_names, 'ridge_coef': ridge.coef_})
lasso_coef = pd.DataFrame({'feature': feature_names, 'lasso_coef': lasso.coef_})

print("Ridge coefficients:")
print(ridge_coef[ridge_coef['ridge_coef'] != 0])

print("\nLasso coefficients (sparsity):")
print(lasso_coef[lasso_coef['lasso_coef'] != 0])
```

## Unsupervised Learning

### K-Means Clustering
```python
from sklearn.cluster import KMeans
from sklearn.metrics import silhouette_score
import matplotlib.pyplot as plt

# Create and fit K-means
kmeans = KMeans(n_clusters=3, random_state=42)
clusters = kmeans.fit_predict(X_scaled)

# Evaluate clustering
silhouette_avg = silhouette_score(X_scaled, clusters)
print(f"Silhouette Score: {silhouette_avg:.3f}")

# Visualize clusters (for 2D data)
plt.scatter(X_scaled[:, 0], X_scaled[:, 1], c=clusters, cmap='viridis')
plt.scatter(kmeans.cluster_centers_[:, 0], kmeans.cluster_centers_[:, 1], s=300, c='red', marker='X')
plt.title('K-Means Clustering')
plt.show()
```

### Hierarchical Clustering
```python
from sklearn.cluster import AgglomerativeClustering
from scipy.cluster.hierarchy import dendrogram, linkage
import matplotlib.pyplot as plt

# Create linkage matrix
linkage_matrix = linkage(X_scaled[:50], method='ward')  # Using subset for visualization

# Plot dendrogram
plt.figure(figsize=(10, 7))
dendrogram(linkage_matrix)
plt.title('Hierarchical Clustering Dendrogram')
plt.show()

# Perform clustering
hierarchical = AgglomerativeClustering(n_clusters=3)
clusters = hierarchical.fit_predict(X_scaled)
```

### Principal Component Analysis (PCA)
```python
from sklearn.decomposition import PCA
import matplotlib.pyplot as plt

# Apply PCA
pca = PCA(n_components=2)
X_pca = pca.fit_transform(X_scaled)

# Explained variance
explained_variance = pca.explained_variance_ratio_
print(f"Explained variance by component: {explained_variance}")
print(f"Total explained variance: {explained_variance.sum():.3f}")

# Visualize PCA
plt.scatter(X_pca[:, 0], X_pca[:, 1], c=y, cmap='viridis')
plt.xlabel('First Principal Component')
plt.ylabel('Second Principal Component')
plt.title('PCA Visualization')
plt.colorbar()
plt.show()
```

## Model Evaluation and Validation

### Cross-Validation
```python
from sklearn.model_selection import cross_val_score, KFold

# K-fold cross-validation
kf = KFold(n_splits=5, shuffle=True, random_state=42)
cv_scores = cross_val_score(model, X, y, cv=kf, scoring='accuracy')

print(f"Cross-validation scores: {cv_scores}")
print(f"Mean CV score: {cv_scores.mean():.3f} (+/- {cv_scores.std() * 2:.3f})")
```

### Hyperparameter Tuning
```python
from sklearn.model_selection import GridSearchCV, RandomizedSearchCV

# Define parameter grid
param_grid = {
    'n_estimators': [50, 100, 200],
    'max_depth': [None, 10, 20, 30],
    'min_samples_split': [2, 5, 10],
    'min_samples_leaf': [1, 2, 4]
}

# Grid search
grid_search = GridSearchCV(
    RandomForestClassifier(random_state=42),
    param_grid,
    cv=5,
    scoring='accuracy',
    n_jobs=-1
)

grid_search.fit(X_train, y_train)

print(f"Best parameters: {grid_search.best_params_}")
print(f"Best cross-validation score: {grid_search.best_score_:.3f}")

# Use best model
best_model = grid_search.best_estimator_
```

### Model Comparison
```python
from sklearn.metrics import accuracy_score, precision_score, recall_score, f1_score
from sklearn.metrics import roc_auc_score, confusion_matrix

models = {
    'KNN': KNeighborsClassifier(),
    'SVM': SVC(probability=True),
    'Decision Tree': DecisionTreeClassifier(),
    'Random Forest': RandomForestClassifier(),
    'Gradient Boosting': GradientBoostingClassifier()
}

results = []

for name, model in models.items():
    model.fit(X_train, y_train)
    y_pred = model.predict(X_test)
    y_prob = model.predict_proba(X_test)[:, 1] if hasattr(model, 'predict_proba') else None

    metrics = {
        'Model': name,
        'Accuracy': accuracy_score(y_test, y_pred),
        'Precision': precision_score(y_test, y_pred),
        'Recall': recall_score(y_test, y_pred),
        'F1-Score': f1_score(y_test, y_pred)
    }

    if y_prob is not None:
        metrics['AUC-ROC'] = roc_auc_score(y_test, y_prob)

    results.append(metrics)

results_df = pd.DataFrame(results)
print(results_df.sort_values('Accuracy', ascending=False))
```

## Real-World Example: Titanic Survival Prediction

Let's build a complete machine learning pipeline for predicting Titanic survival.

### Data Loading and Exploration
```python
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler, LabelEncoder
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import accuracy_score, classification_report

# Load data
df = pd.read_csv('titanic.csv')

# Basic exploration
print(df.head())
print(df.info())
print(df.describe())

# Check missing values
print(df.isnull().sum())
```

### Data Preprocessing
```python
# Handle missing values
df['Age'].fillna(df['Age'].median(), inplace=True)
df['Embarked'].fillna(df['Embarked'].mode()[0], inplace=True)
df.drop(['Cabin', 'Name', 'Ticket', 'PassengerId'], axis=1, inplace=True)

# Encode categorical variables
le = LabelEncoder()
df['Sex'] = le.fit_transform(df['Sex'])
df['Embarked'] = le.fit_transform(df['Embarked'])

# Create new features
df['FamilySize'] = df['SibSp'] + df['Parch'] + 1
df['IsAlone'] = (df['FamilySize'] == 1).astype(int)

# Prepare features and target
X = df.drop('Survived', axis=1)
y = df['Survived']

# Split data
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

# Scale features
scaler = StandardScaler()
X_train_scaled = scaler.fit_transform(X_train)
X_test_scaled = scaler.transform(X_test)
```

### Model Training and Evaluation
```python
# Train Random Forest model
rf_model = RandomForestClassifier(n_estimators=100, random_state=42)
rf_model.fit(X_train_scaled, y_train)

# Make predictions
y_pred = rf_model.predict(X_test_scaled)
y_prob = rf_model.predict_proba(X_test_scaled)[:, 1]

# Evaluate model
accuracy = accuracy_score(y_test, y_pred)
print(f"Accuracy: {accuracy:.3f}")

print("\nClassification Report:")
print(classification_report(y_test, y_pred))

# Feature importance
feature_importance = pd.DataFrame({
    'feature': X.columns,
    'importance': rf_model.feature_importances_
}).sort_values('importance', ascending=False)

print("\nFeature Importance:")
print(feature_importance)
```

### Model Interpretation
```python
# Plot feature importance
plt.figure(figsize=(10, 6))
sns.barplot(x='importance', y='feature', data=feature_importance)
plt.title('Feature Importance in Titanic Survival Prediction')
plt.show()

# Confusion matrix
from sklearn.metrics import confusion_matrix
import seaborn as sns

cm = confusion_matrix(y_test, y_pred)
plt.figure(figsize=(8, 6))
sns.heatmap(cm, annot=True, fmt='d', cmap='Blues')
plt.title('Confusion Matrix')
plt.ylabel('True Label')
plt.xlabel('Predicted Label')
plt.show()
```

## Best Practices

### Data Science Workflow
1. **Problem Definition**: Clearly define the problem and success criteria
2. **Data Collection**: Gather relevant data from various sources
3. **Data Cleaning**: Handle missing values, outliers, and inconsistencies
4. **Exploratory Analysis**: Understand data distributions and relationships
5. **Feature Engineering**: Create meaningful features from raw data
6. **Model Selection**: Choose appropriate algorithms for the problem
7. **Model Training**: Train models using cross-validation
8. **Hyperparameter Tuning**: Optimize model parameters
9. **Model Evaluation**: Assess performance using appropriate metrics
10. **Model Deployment**: Deploy model to production environment

### Avoiding Common Pitfalls
- **Data Leakage**: Ensure no information from test set leaks into training
- **Overfitting**: Use regularization and cross-validation
- **Underfitting**: Ensure model complexity matches data complexity
- **Imbalanced Data**: Handle class imbalance appropriately
- **Feature Scaling**: Scale features when using distance-based algorithms

### Performance Metrics
- **Classification**: Accuracy, Precision, Recall, F1-Score, AUC-ROC
- **Regression**: MSE, RMSE, MAE, R²
- **Clustering**: Silhouette Score, Calinski-Harabasz Index

## Next Steps

You've learned the fundamentals of machine learning! Here's what to explore next:

1. **Deep Learning**: Neural networks with TensorFlow or PyTorch
2. **Natural Language Processing**: Text analysis and language models
3. **Computer Vision**: Image recognition and processing
4. **Time Series Analysis**: Forecasting and temporal data
5. **Reinforcement Learning**: Agent-based learning
6. **MLOps**: Model deployment and monitoring
7. **Ethics in ML**: Bias, fairness, and responsible AI

Practice with datasets from Kaggle, implement algorithms from scratch, and contribute to open-source ML projects. Machine learning is a rapidly evolving field - stay curious and keep learning!