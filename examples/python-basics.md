# Python Programming Basics

Learn the fundamentals of Python programming in this comprehensive beginner-friendly course. No prior programming experience required.

## What is Python?

Python is a high-level, interpreted programming language known for its simplicity and readability. Created by Guido van Rossum and first released in 1991, Python emphasizes code readability and allows programmers to express concepts in fewer lines of code.

### Key Features
- **Simple syntax** similar to English
- **Dynamic typing** - no need to declare variable types
- **Cross-platform** compatibility
- **Extensive libraries** for various tasks
- **Large community** support

## Installing Python

### Windows
1. Visit python.org/downloads
2. Download the latest Python installer
3. Run the installer
4. Check "Add Python to PATH"
5. Complete installation

### macOS
```bash
# Using Homebrew
brew install python

# Or download from python.org
```

### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install python3 python3-pip
```

### Verification
Open terminal/command prompt and type:
```bash
python --version
# or
python3 --version
```

## Your First Python Program

Let's create the traditional "Hello, World!" program.

### Using Python Interpreter
```bash
python3
>>> print("Hello, World!")
Hello, World!
>>> exit()
```

### Creating a Script File
Create a file named `hello.py`:

```python
# This is a comment
print("Hello, World!")
```

Run the script:
```bash
python hello.py
# Output: Hello, World!
```

## Variables and Data Types

Python variables don't need explicit type declarations.

### Basic Data Types
```python
# Numbers
age = 25          # Integer
height = 5.9      # Float
temperature = -5  # Negative number

# Strings
name = "Alice"
message = 'Hello, World!'
multiline = """This is a
multiline string"""

# Booleans
is_student = True
has_job = False
```

### Variable Naming Rules
- Start with letter or underscore
- Can contain letters, numbers, underscores
- Case-sensitive
- No reserved keywords

## Basic Operations

### Arithmetic Operations
```python
# Addition
result = 5 + 3  # 8

# Subtraction
result = 10 - 4  # 6

# Multiplication
result = 6 * 7  # 42

# Division
result = 15 / 3  # 5.0 (float division)
result = 15 // 3 # 5 (integer division)

# Modulo
result = 17 % 5  # 2

# Exponentiation
result = 2 ** 3  # 8
```

### String Operations
```python
# Concatenation
full_name = "John" + " " + "Doe"  # "John Doe"

# Repetition
laugh = "Ha" * 3  # "HaHaHa"

# Length
name = "Python"
length = len(name)  # 6

# Indexing
first_char = name[0]  # 'P'
last_char = name[-1]  # 'n'
```

## Control Structures

### Conditional Statements
```python
age = 18

if age >= 18:
    print("You are an adult")
elif age >= 13:
    print("You are a teenager")
else:
    print("You are a child")

# Ternary operator
status = "adult" if age >= 18 else "minor"
```

### Loops

#### For Loops
```python
# Iterate over a list
fruits = ["apple", "banana", "orange"]
for fruit in fruits:
    print(f"I like {fruit}")

# Iterate with range
for i in range(5):  # 0, 1, 2, 3, 4
    print(f"Count: {i}")

# Iterate with index
for index, fruit in enumerate(fruits):
    print(f"{index}: {fruit}")
```

#### While Loops
```python
count = 0
while count < 5:
    print(f"Count: {count}")
    count += 1

# Infinite loop with break
while True:
    user_input = input("Enter 'quit' to exit: ")
    if user_input == 'quit':
        break
    print(f"You entered: {user_input}")
```

## Functions

Functions are reusable blocks of code.

### Defining Functions
```python
def greet(name):
    """This function greets a person by name."""
    return f"Hello, {name}!"

# Call the function
message = greet("Alice")
print(message)  # "Hello, Alice!"

# Function with default parameter
def greet_with_time(name, time="morning"):
    return f"Good {time}, {name}!"

print(greet_with_time("Bob"))           # "Good morning, Bob!"
print(greet_with_time("Bob", "evening")) # "Good evening, Bob!"
```

### Lambda Functions
```python
# Simple lambda function
square = lambda x: x ** 2
print(square(5))  # 25

# Lambda with multiple parameters
add = lambda x, y: x + y
print(add(3, 4))  # 7
```

## Data Structures

### Lists
```python
# Creating lists
numbers = [1, 2, 3, 4, 5]
fruits = ["apple", "banana", "orange"]
mixed = [1, "hello", 3.14, True]

# Accessing elements
first = numbers[0]    # 1
last = numbers[-1]    # 5

# Slicing
first_three = numbers[:3]  # [1, 2, 3]
middle = numbers[1:4]      # [2, 3, 4]

# Modifying lists
numbers.append(6)      # [1, 2, 3, 4, 5, 6]
numbers.insert(0, 0)   # [0, 1, 2, 3, 4, 5, 6]
numbers.remove(3)      # [0, 1, 2, 4, 5, 6]
popped = numbers.pop() # Removes and returns last element

# List methods
length = len(numbers)      # 6
total = sum(numbers)       # 17
maximum = max(numbers)     # 6
minimum = min(numbers)     # 0
```

### Tuples
```python
# Creating tuples (immutable)
coordinates = (10, 20)
person = ("Alice", 25, "Engineer")

# Accessing elements
x, y = coordinates  # Unpacking
name, age, job = person

# Tuples are immutable
# coordinates[0] = 15  # This would raise an error
```

### Dictionaries
```python
# Creating dictionaries
person = {
    "name": "Alice",
    "age": 25,
    "city": "New York",
    "job": "Engineer"
}

# Accessing values
name = person["name"]      # "Alice"
age = person.get("age")    # 25
country = person.get("country", "USA")  # "USA" (default)

# Modifying dictionaries
person["age"] = 26         # Update value
person["email"] = "alice@example.com"  # Add new key-value pair
del person["city"]          # Remove key-value pair

# Dictionary methods
keys = person.keys()        # dict_keys(['name', 'age', 'job', 'email'])
values = person.values()    # dict_values(['Alice', 26, 'Engineer', 'alice@example.com'])
items = person.items()      # dict_items([('name', 'Alice'), ('age', 26), ...])
```

## File Handling

### Reading Files
```python
# Reading entire file
with open("example.txt", "r") as file:
    content = file.read()
    print(content)

# Reading line by line
with open("example.txt", "r") as file:
    for line in file:
        print(line.strip())

# Reading all lines into a list
with open("example.txt", "r") as file:
    lines = file.readlines()
    print(lines)
```

### Writing Files
```python
# Writing to a file (overwrites existing content)
with open("output.txt", "w") as file:
    file.write("Hello, World!\n")
    file.write("This is a new line.\n")

# Appending to a file
with open("output.txt", "a") as file:
    file.write("This line will be appended.\n")

# Writing multiple lines
lines = ["Line 1", "Line 2", "Line 3"]
with open("output.txt", "w") as file:
    file.writelines(line + "\n" for line in lines)
```

## Error Handling

Python uses try-except blocks for error handling.

```python
try:
    # Code that might raise an exception
    result = 10 / 0
except ZeroDivisionError:
    print("Cannot divide by zero!")
except Exception as e:
    print(f"An error occurred: {e}")
else:
    print("No errors occurred")
finally:
    print("This always executes")

# Raising exceptions
def divide(a, b):
    if b == 0:
        raise ValueError("Cannot divide by zero")
    return a / b

try:
    result = divide(10, 0)
except ValueError as e:
    print(e)  # "Cannot divide by zero"
```

## Modules and Packages

### Importing Modules
```python
# Import entire module
import math
print(math.sqrt(16))  # 4.0

# Import specific functions
from math import sqrt, pi
print(sqrt(25))  # 5.0
print(pi)        # 3.141592653589793

# Import with alias
import math as m
print(m.cos(0))  # 1.0

# Import all (not recommended)
from math import *
print(sin(0))    # 0.0
```

### Creating Your Own Module
Create a file called `mymodule.py`:

```python
def greet(name):
    return f"Hello, {name}!"

PI = 3.14159

class Calculator:
    def add(self, a, b):
        return a + b

    def multiply(self, a, b):
        return a * b
```

Use the module:
```python
import mymodule

print(mymodule.greet("Alice"))  # "Hello, Alice!"
print(mymodule.PI)              # 3.14159

calc = mymodule.Calculator()
print(calc.add(5, 3))           # 8
```

## Classes and Objects

Python supports object-oriented programming.

### Defining Classes
```python
class Person:
    # Class attribute
    species = "Human"

    # Constructor
    def __init__(self, name, age):
        self.name = name  # Instance attribute
        self.age = age

    # Instance method
    def greet(self):
        return f"Hello, my name is {self.name}"

    # Method with parameters
    def celebrate_birthday(self, years=1):
        self.age += years
        return f"Happy birthday! You are now {self.age} years old"

    # Class method
    @classmethod
    def create_anonymous(cls):
        return cls("Anonymous", 0)

    # Static method
    @staticmethod
    def is_adult(age):
        return age >= 18
```

### Using Classes
```python
# Create instances
alice = Person("Alice", 25)
bob = Person("Bob", 30)

# Access attributes
print(alice.name)    # "Alice"
print(alice.age)     # 25

# Call methods
print(alice.greet())  # "Hello, my name is Alice"
print(alice.celebrate_birthday())  # "Happy birthday! You are now 26 years old"

# Class method
anonymous = Person.create_anonymous()
print(anonymous.name)  # "Anonymous"

# Static method
print(Person.is_adult(20))  # True
print(Person.is_adult(15))  # False
```

## Next Steps

Congratulations! You've learned the basics of Python programming. Here are some recommendations for continuing your journey:

1. **Practice regularly** - Write small programs daily
2. **Work on projects** - Build something you're interested in
3. **Read code** - Study open-source Python projects
4. **Join communities** - Participate in Python forums and meetups
5. **Learn frameworks** - Explore Django for web development, Flask for APIs, or Pandas for data analysis

Remember, programming is a skill that improves with practice. Keep coding and have fun!