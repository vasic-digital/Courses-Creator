# Data Analysis with Python

Learn to analyze and visualize data using Python's powerful data science ecosystem. This course covers pandas, NumPy, and matplotlib for data manipulation and visualization.

## Introduction to Data Analysis

Data analysis is the process of inspecting, cleaning, transforming, and modeling data to discover useful information, draw conclusions, and support decision-making.

### Why Python for Data Analysis?
- **Rich Ecosystem**: Extensive libraries for data manipulation and analysis
- **Easy to Learn**: Simple syntax, readable code
- **Open Source**: Free and community-driven
- **Versatile**: Used in academia, industry, and research
- **Integration**: Works well with other languages and tools

## Setting Up Your Environment

### Installing Python Data Science Stack

#### Using Anaconda (Recommended)
```bash
# Download and install Anaconda from anaconda.com

# Create a new environment
conda create -n data-analysis python=3.9

# Activate the environment
conda activate data-analysis

# Install core packages
conda install numpy pandas matplotlib seaborn jupyter
```

#### Using pip
```bash
# Install core packages
pip install numpy pandas matplotlib seaborn jupyter

# Optional: Install Jupyter notebook
pip install notebook
```

### Jupyter Notebook
```bash
# Start Jupyter
jupyter notebook

# Or JupyterLab (more modern interface)
pip install jupyterlab
jupyter lab
```

## NumPy: Numerical Computing

NumPy provides support for large, multi-dimensional arrays and matrices, along with mathematical functions to operate on them.

### Creating Arrays
```python
import numpy as np

# Create arrays from lists
arr1d = np.array([1, 2, 3, 4, 5])
arr2d = np.array([[1, 2, 3], [4, 5, 6]])

# Special arrays
zeros = np.zeros((3, 4))        # 3x4 array of zeros
ones = np.ones((2, 3))          # 2x3 array of ones
identity = np.eye(3)            # 3x3 identity matrix
random_arr = np.random.rand(2, 3)  # Random array

# Ranges
range_arr = np.arange(10)       # [0, 1, 2, ..., 9]
linspace_arr = np.linspace(0, 1, 5)  # [0, 0.25, 0.5, 0.75, 1]
```

### Array Operations
```python
a = np.array([1, 2, 3])
b = np.array([4, 5, 6])

# Element-wise operations
print(a + b)  # [5, 7, 9]
print(a * b)  # [4, 10, 18]
print(a ** 2) # [1, 4, 9]

# Mathematical functions
print(np.sin(a))    # Sine of each element
print(np.sqrt(a))   # Square root of each element
print(np.exp(a))    # Exponential of each element

# Aggregation
print(np.sum(a))    # Sum: 6
print(np.mean(a))   # Mean: 2.0
print(np.std(a))    # Standard deviation
print(np.min(a))    # Minimum: 1
print(np.max(a))    # Maximum: 3
```

### Indexing and Slicing
```python
arr = np.array([[1, 2, 3, 4],
                [5, 6, 7, 8],
                [9, 10, 11, 12]])

# Basic indexing
print(arr[0, 0])   # 1 (first row, first column)
print(arr[1, 2])   # 7 (second row, third column)

# Row slicing
print(arr[0, :])   # [1, 2, 3, 4] (first row)
print(arr[:, 1])   # [2, 6, 10] (second column)

# Subarray slicing
print(arr[0:2, 1:3])  # [[2, 3], [6, 7]]

# Boolean indexing
mask = arr > 5
print(arr[mask])    # [6, 7, 8, 9, 10, 11, 12]
```

### Array Manipulation
```python
arr = np.array([[1, 2], [3, 4]])

# Reshaping
reshaped = arr.reshape(4, 1)  # 4x1 array
flattened = arr.flatten()      # 1D array

# Transposing
transposed = arr.T

# Concatenation
a = np.array([[1, 2], [3, 4]])
b = np.array([[5, 6], [7, 8]])

vertical = np.vstack((a, b))   # Stack vertically
horizontal = np.hstack((a, b)) # Stack horizontally

# Splitting
left, right = np.hsplit(arr, 2)    # Split into 2 horizontally
top, bottom = np.vsplit(arr, 2)    # Split into 2 vertically
```

## Pandas: Data Manipulation

Pandas provides data structures and operations for working with structured data.

### Series and DataFrame
```python
import pandas as pd

# Series (1D labeled array)
s = pd.Series([1, 3, 5, 6, 8])
s_with_index = pd.Series([1, 3, 5], index=['a', 'b', 'c'])

# DataFrame (2D labeled data structure)
data = {
    'Name': ['Alice', 'Bob', 'Charlie', 'Diana'],
    'Age': [25, 30, 35, 28],
    'City': ['NYC', 'LA', 'Chicago', 'Houston']
}

df = pd.DataFrame(data)

# Display DataFrame
print(df)
```

### Reading Data
```python
# CSV files
df_csv = pd.read_csv('data.csv')

# Excel files
df_excel = pd.read_excel('data.xlsx')

# JSON files
df_json = pd.read_json('data.json')

# SQL databases
import sqlite3
conn = sqlite3.connect('database.db')
df_sql = pd.read_sql('SELECT * FROM table_name', conn)

# From dictionaries
df_dict = pd.DataFrame.from_dict(data_dict)
```

### Data Exploration
```python
# Basic information
print(df.head())     # First 5 rows
print(df.tail())     # Last 5 rows
print(df.info())     # Data types and non-null counts
print(df.describe()) # Statistical summary

# Shape and size
print(df.shape)      # (rows, columns)
print(df.columns)    # Column names
print(df.index)      # Index

# Data types
print(df.dtypes)

# Missing values
print(df.isnull().sum())  # Count missing values per column
```

### Selecting Data
```python
# Column selection
names = df['Name']           # Single column (Series)
subset = df[['Name', 'Age']] # Multiple columns (DataFrame)

# Row selection
first_row = df.iloc[0]       # By position
second_row = df.loc[1]       # By label

# Conditional selection
adults = df[df['Age'] >= 30]
nyc_people = df[df['City'] == 'NYC']

# Multiple conditions
young_nyc = df[(df['Age'] < 30) & (df['City'] == 'NYC')]
```

### Data Cleaning
```python
# Handling missing values
df.dropna()                    # Remove rows with NaN
df.fillna(0)                   # Fill NaN with 0
df.fillna(df.mean())           # Fill with column mean

# Removing duplicates
df.drop_duplicates()

# Data type conversion
df['Age'] = df['Age'].astype(int)
df['Date'] = pd.to_datetime(df['Date'])

# String operations
df['Name'] = df['Name'].str.upper()
df['Name'] = df['Name'].str.strip()
```

### Data Transformation
```python
# Adding new columns
df['Age_Group'] = pd.cut(df['Age'], bins=[0, 18, 65, 100], labels=['Child', 'Adult', 'Senior'])
df['Full_Name'] = df['First_Name'] + ' ' + df['Last_Name']

# Applying functions
df['Age_Squared'] = df['Age'].apply(lambda x: x ** 2)
df['Name_Length'] = df['Name'].apply(len)

# Grouping and aggregation
grouped = df.groupby('City')
city_stats = grouped['Age'].agg(['mean', 'min', 'max', 'count'])

# Pivot tables
pivot = df.pivot_table(values='Age', index='City', columns='Department', aggfunc='mean')
```

### Merging and Joining
```python
# Sample data
employees = pd.DataFrame({
    'ID': [1, 2, 3, 4],
    'Name': ['Alice', 'Bob', 'Charlie', 'Diana'],
    'Dept_ID': [101, 102, 101, 103]
})

departments = pd.DataFrame({
    'Dept_ID': [101, 102, 103],
    'Dept_Name': ['Engineering', 'Sales', 'Marketing']
})

# Inner join (default)
result = pd.merge(employees, departments, on='Dept_ID')

# Left join
result = pd.merge(employees, departments, on='Dept_ID', how='left')

# Outer join
result = pd.merge(employees, departments, on='Dept_ID', how='outer')

# Concatenation
df1 = pd.DataFrame({'A': [1, 2], 'B': [3, 4]})
df2 = pd.DataFrame({'A': [5, 6], 'B': [7, 8]})
result = pd.concat([df1, df2])
```

## Matplotlib: Data Visualization

Matplotlib is a comprehensive library for creating static, animated, and interactive visualizations.

### Basic Plotting
```python
import matplotlib.pyplot as plt

# Line plot
x = [1, 2, 3, 4, 5]
y = [2, 4, 6, 8, 10]

plt.plot(x, y)
plt.title('Simple Line Plot')
plt.xlabel('X values')
plt.ylabel('Y values')
plt.show()

# Scatter plot
plt.scatter(x, y)
plt.title('Scatter Plot')
plt.show()
```

### Multiple Plots
```python
# Multiple lines
x = [1, 2, 3, 4, 5]
y1 = [1, 4, 9, 16, 25]
y2 = [1, 8, 27, 64, 125]

plt.plot(x, y1, label='x²')
plt.plot(x, y2, label='x³')
plt.legend()
plt.title('Multiple Lines')
plt.show()
```

### Bar Charts
```python
categories = ['A', 'B', 'C', 'D']
values = [23, 45, 56, 78]

plt.bar(categories, values)
plt.title('Bar Chart')
plt.xlabel('Categories')
plt.ylabel('Values')
plt.show()

# Horizontal bar chart
plt.barh(categories, values)
plt.title('Horizontal Bar Chart')
plt.show()
```

### Histograms
```python
import numpy as np

# Generate random data
data = np.random.normal(0, 1, 1000)

plt.hist(data, bins=30, alpha=0.7)
plt.title('Histogram')
plt.xlabel('Value')
plt.ylabel('Frequency')
plt.show()
```

### Pie Charts
```python
labels = ['Apples', 'Bananas', 'Cherries', 'Dates']
sizes = [30, 25, 20, 25]
colors = ['red', 'yellow', 'purple', 'brown']

plt.pie(sizes, labels=labels, colors=colors, autopct='%1.1f%%')
plt.title('Fruit Distribution')
plt.axis('equal')  # Equal aspect ratio ensures pie is drawn as a circle
plt.show()
```

### Subplots
```python
# Create figure with subplots
fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(10, 8))

# Plot on each subplot
ax1.plot(x, y1)
ax1.set_title('Line Plot')

ax2.scatter(x, y2)
ax2.set_title('Scatter Plot')

ax3.bar(categories, values)
ax3.set_title('Bar Chart')

ax4.hist(data, bins=20)
ax4.set_title('Histogram')

plt.tight_layout()
plt.show()
```

### Customizing Plots
```python
# Colors, markers, line styles
plt.plot(x, y, color='red', marker='o', linestyle='--', linewidth=2, markersize=8)

# Adding grid
plt.grid(True, alpha=0.3)

# Setting limits
plt.xlim(0, 6)
plt.ylim(0, 12)

# Adding text
plt.text(3, 8, 'Important Point', fontsize=12, ha='center')

# Saving plots
plt.savefig('my_plot.png', dpi=300, bbox_inches='tight')
```

## Seaborn: Statistical Visualization

Seaborn provides a high-level interface for drawing attractive statistical graphics.

### Basic Plots
```python
import seaborn as sns
import matplotlib.pyplot as plt

# Set style
sns.set_style("whitegrid")

# Sample data
tips = sns.load_dataset("tips")

# Scatter plot with regression line
sns.regplot(x="total_bill", y="tip", data=tips)
plt.title('Tips vs Total Bill')
plt.show()

# Categorical scatter plot
sns.stripplot(x="day", y="total_bill", data=tips)
plt.title('Total Bill by Day')
plt.show()
```

### Distribution Plots
```python
# Histogram with KDE
sns.histplot(data=tips, x="total_bill", kde=True)
plt.title('Distribution of Total Bills')
plt.show()

# Box plot
sns.boxplot(x="day", y="total_bill", data=tips)
plt.title('Total Bill Distribution by Day')
plt.show()

# Violin plot
sns.violinplot(x="day", y="total_bill", data=tips)
plt.title('Total Bill Distribution by Day')
plt.show()
```

### Categorical Plots
```python
# Bar plot
sns.barplot(x="day", y="total_bill", data=tips)
plt.title('Average Total Bill by Day')
plt.show()

# Count plot
sns.countplot(x="day", data=tips)
plt.title('Number of Observations by Day')
plt.show()
```

### Matrix Plots
```python
# Correlation heatmap
numeric_cols = tips.select_dtypes(include=['float64', 'int64'])
correlation = numeric_cols.corr()

sns.heatmap(correlation, annot=True, cmap='coolwarm')
plt.title('Correlation Heatmap')
plt.show()

# Pair plot
sns.pairplot(tips, hue="sex")
plt.show()
```

## Real-World Data Analysis Example

Let's analyze a dataset step by step.

### Loading and Exploring Data
```python
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns

# Load sample dataset
df = pd.read_csv('https://raw.githubusercontent.com/mwaskom/seaborn-data/master/titanic.csv')

# Basic exploration
print(df.head())
print(df.info())
print(df.describe())

# Check for missing values
print(df.isnull().sum())
```

### Data Cleaning
```python
# Handle missing values
df['age'].fillna(df['age'].median(), inplace=True)
df['embarked'].fillna(df['embarked'].mode()[0], inplace=True)
df.drop('deck', axis=1, inplace=True)  # Too many missing values

# Convert data types
df['survived'] = df['survived'].astype(bool)
df['pclass'] = df['pclass'].astype('category')

# Create new features
df['family_size'] = df['sibsp'] + df['parch'] + 1
df['is_alone'] = df['family_size'] == 1
```

### Exploratory Data Analysis
```python
# Survival rate by class
survival_by_class = df.groupby('pclass')['survived'].mean()
print(survival_by_class)

# Age distribution
plt.figure(figsize=(10, 6))
sns.histplot(data=df, x='age', hue='survived', multiple='stack')
plt.title('Age Distribution by Survival')
plt.show()

# Survival by gender and class
plt.figure(figsize=(10, 6))
sns.barplot(x='pclass', y='survived', hue='sex', data=df)
plt.title('Survival Rate by Class and Gender')
plt.show()

# Correlation analysis
numeric_cols = df.select_dtypes(include=[np.number])
correlation = numeric_cols.corr()

plt.figure(figsize=(10, 8))
sns.heatmap(correlation, annot=True, cmap='coolwarm')
plt.title('Correlation Matrix')
plt.show()
```

### Statistical Analysis
```python
from scipy import stats

# T-test: Age difference between survivors and non-survivors
survivors_age = df[df['survived']]['age']
non_survivors_age = df[~df['survived']]['age']

t_stat, p_value = stats.ttest_ind(survivors_age, non_survivors_age)
print(f"T-statistic: {t_stat:.3f}, P-value: {p_value:.3f}")

# Chi-square test: Relationship between class and survival
contingency_table = pd.crosstab(df['pclass'], df['survived'])
chi2, p, dof, expected = stats.chi2_contingency(contingency_table)
print(f"Chi-square: {chi2:.3f}, P-value: {p:.3f}")
```

## Best Practices

### Code Organization
```
project/
├── data/
│   ├── raw/
│   ├── processed/
│   └── external/
├── notebooks/
├── src/
│   ├── data/
│   ├── features/
│   ├── models/
│   └── visualization/
├── tests/
└── reports/
```

### Data Analysis Workflow
1. **Define the problem** - What question are you trying to answer?
2. **Collect data** - Gather relevant data from various sources
3. **Clean data** - Handle missing values, outliers, inconsistencies
4. **Explore data** - Understand distributions, relationships, patterns
5. **Analyze data** - Apply statistical methods and modeling
6. **Visualize results** - Create clear, informative visualizations
7. **Communicate findings** - Present insights to stakeholders

### Performance Tips
- Use vectorized operations in NumPy and pandas
- Avoid loops when possible
- Use appropriate data types
- Consider memory usage for large datasets
- Use sampling for exploratory analysis

## Next Steps

You've learned the fundamentals of data analysis with Python! Here's what to explore next:

1. **Machine Learning**: Scikit-learn for predictive modeling
2. **Big Data**: PySpark for large-scale data processing
3. **Web Scraping**: Beautiful Soup and Scrapy for data collection
4. **Time Series**: Analysis of temporal data
5. **Geospatial Analysis**: Working with location data
6. **Deep Learning**: TensorFlow or PyTorch for neural networks
7. **Data Engineering**: Building data pipelines and ETL processes

Practice with real datasets from Kaggle, UCI Machine Learning Repository, or government open data portals. The key to becoming proficient in data analysis is consistent practice and curiosity about data!