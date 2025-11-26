# JavaScript Essentials

Master the core concepts of JavaScript programming. This course covers everything from basic syntax to advanced features in modern JavaScript.

## Introduction to JavaScript

JavaScript is a versatile programming language primarily used for web development. Created by Brendan Eich in 1995, it has evolved from a simple scripting language into a powerful, full-featured programming language.

### What Makes JavaScript Special?
- **Dynamic typing** - Variables can hold any type
- **First-class functions** - Functions are treated as values
- **Prototype-based inheritance** - Objects inherit from other objects
- **Event-driven** - Responds to user interactions
- **Cross-platform** - Runs in browsers and servers (Node.js)

## Setting Up Your Environment

### Browser Console
The easiest way to start with JavaScript is using your browser's developer console.

**Chrome/Edge**: Press F12 → Console tab
**Firefox**: Press F12 → Console tab
**Safari**: Develop menu → Show Web Inspector → Console

Try this in the console:
```javascript
console.log("Hello, JavaScript!");
```

### Code Editor
For larger projects, use a code editor like:
- Visual Studio Code (recommended)
- Sublime Text
- Atom

### Node.js (Optional)
For server-side JavaScript, install Node.js from nodejs.org

## Variables and Data Types

JavaScript has dynamic typing - you don't need to declare variable types.

### Variable Declaration
```javascript
// var (older way, avoid in modern code)
var name = "Alice";

// let (block-scoped, can be reassigned)
let age = 25;
age = 26; // OK

// const (block-scoped, cannot be reassigned)
const PI = 3.14159;
// PI = 3.14; // Error!
```

### Primitive Data Types
```javascript
// Numbers
let integer = 42;
let float = 3.14;
let negative = -10;

// Strings
let singleQuotes = 'Hello';
let doubleQuotes = "World";
let templateLiteral = `Hello, ${name}!`; // "Hello, Alice!"

// Booleans
let isStudent = true;
let hasJob = false;

// Undefined and Null
let undefinedVar; // undefined
let emptyValue = null; // intentional empty value

// Symbols (unique identifiers)
let uniqueId = Symbol('id');

// BigInt (for large numbers)
let bigNumber = 123456789012345678901234567890n;
```

## Operators

### Arithmetic Operators
```javascript
let a = 10;
let b = 3;

console.log(a + b);  // 13 (addition)
console.log(a - b);  // 7 (subtraction)
console.log(a * b);  // 30 (multiplication)
console.log(a / b);  // 3.333... (division)
console.log(a % b);  // 1 (modulo)
console.log(a ** b); // 1000 (exponentiation)
```

### Comparison Operators
```javascript
console.log(5 == '5');   // true (loose equality)
console.log(5 === '5');  // false (strict equality)
console.log(5 != '5');   // false (loose inequality)
console.log(5 !== '5');  // true (strict inequality)

console.log(10 > 5);     // true
console.log(10 < 5);     // false
console.log(10 >= 10);   // true
console.log(5 <= 3);     // false
```

### Logical Operators
```javascript
let x = true;
let y = false;

console.log(x && y);  // false (AND)
console.log(x || y);  // true (OR)
console.log(!x);      // false (NOT)
```

## Control Structures

### Conditional Statements
```javascript
let age = 18;

if (age >= 18) {
    console.log("You can vote!");
} else if (age >= 16) {
    console.log("You can drive!");
} else {
    console.log("You're too young for both.");
}

// Ternary operator
let canVote = age >= 18 ? "Yes" : "No";
console.log(canVote);

// Switch statement
let day = "Monday";
switch (day) {
    case "Monday":
        console.log("Start of work week");
        break;
    case "Friday":
        console.log("TGIF!");
        break;
    default:
        console.log("Regular day");
}
```

### Loops

#### For Loop
```javascript
// Traditional for loop
for (let i = 0; i < 5; i++) {
    console.log(`Count: ${i}`);
}

// For...of loop (for arrays)
let fruits = ['apple', 'banana', 'orange'];
for (let fruit of fruits) {
    console.log(fruit);
}

// For...in loop (for objects)
let person = { name: 'Alice', age: 25 };
for (let key in person) {
    console.log(`${key}: ${person[key]}`);
}
```

#### While Loops
```javascript
// While loop
let count = 0;
while (count < 5) {
    console.log(count);
    count++;
}

// Do...while loop (executes at least once)
let number;
do {
    number = Math.floor(Math.random() * 10);
    console.log(`Generated: ${number}`);
} while (number !== 5);
```

## Functions

Functions are first-class citizens in JavaScript.

### Function Declaration
```javascript
// Function declaration
function greet(name) {
    return `Hello, ${name}!`;
}

console.log(greet("Alice")); // "Hello, Alice!"

// Function with default parameters
function createUser(name, age = 18) {
    return { name, age };
}

console.log(createUser("Bob")); // { name: "Bob", age: 18 }
```

### Function Expression
```javascript
// Anonymous function expression
const add = function(a, b) {
    return a + b;
};

// Named function expression
const multiply = function multiplyNumbers(x, y) {
    return x * y;
};
```

### Arrow Functions (ES6+)
```javascript
// Basic arrow function
const square = x => x * x;
console.log(square(5)); // 25

// Arrow function with multiple parameters
const add = (a, b) => a + b;
console.log(add(3, 4)); // 7

// Arrow function with body
const greet = name => {
    const message = `Hello, ${name}!`;
    return message.toUpperCase();
};
console.log(greet("Alice")); // "HELLO, ALICE!"

// Returning objects
const createPerson = (name, age) => ({ name, age });
console.log(createPerson("Bob", 25)); // { name: "Bob", age: 25 }
```

## Arrays

Arrays are dynamic, ordered collections of values.

### Creating Arrays
```javascript
// Array literal
let numbers = [1, 2, 3, 4, 5];

// Mixed types
let mixed = [1, "hello", true, null];

// Empty array
let empty = [];

// Array constructor (less common)
let constructed = new Array(5); // Creates array with 5 empty slots
```

### Array Methods
```javascript
let fruits = ['apple', 'banana'];

// Adding elements
fruits.push('orange');        // ['apple', 'banana', 'orange']
fruits.unshift('grape');      // ['grape', 'apple', 'banana', 'orange']

// Removing elements
let last = fruits.pop();      // 'orange', fruits = ['grape', 'apple', 'banana']
let first = fruits.shift();   // 'grape', fruits = ['apple', 'banana']

// Finding elements
let index = fruits.indexOf('banana'); // 1
let hasApple = fruits.includes('apple'); // true

// Slicing and splicing
let sliced = fruits.slice(1, 3); // ['banana'] (doesn't modify original)
let removed = fruits.splice(1, 1, 'cherry'); // fruits = ['apple', 'cherry']

// Iterating
fruits.forEach(fruit => console.log(fruit));

// Transforming
let upperFruits = fruits.map(fruit => fruit.toUpperCase());
let longFruits = fruits.filter(fruit => fruit.length > 5);

// Reducing
let totalLength = fruits.reduce((sum, fruit) => sum + fruit.length, 0);
```

## Objects

Objects are collections of key-value pairs.

### Creating Objects
```javascript
// Object literal
let person = {
    name: "Alice",
    age: 25,
    city: "New York",
    greet: function() {
        return `Hello, I'm ${this.name}`;
    }
};

// Empty object
let emptyObj = {};

// Constructor
let constructed = new Object();
```

### Accessing Properties
```javascript
// Dot notation
console.log(person.name);    // "Alice"
console.log(person.age);     // 25

// Bracket notation
console.log(person['city']); // "New York"

// Dynamic property access
let property = 'name';
console.log(person[property]); // "Alice"
```

### Modifying Objects
```javascript
// Adding properties
person.job = "Developer";
person['salary'] = 75000;

// Updating properties
person.age = 26;

// Deleting properties
delete person.city;

// Checking properties
console.log('name' in person);        // true
console.log(person.hasOwnProperty('age')); // true
```

### Object Methods
```javascript
let user = {
    name: "Alice",
    age: 25,
    hobbies: ["reading", "coding"]
};

// Getting keys, values, entries
console.log(Object.keys(user));    // ["name", "age", "hobbies"]
console.log(Object.values(user));  // ["Alice", 25, ["reading", "coding"]]
console.log(Object.entries(user)); // [["name", "Alice"], ["age", 25], ...]

// Cloning objects
let shallowCopy = Object.assign({}, user);
let spreadCopy = { ...user };

// Preventing modifications
Object.freeze(user);  // Cannot add, delete, or modify properties
Object.seal(user);    // Cannot add or delete properties, but can modify
```

## Classes (ES6+)

JavaScript supports class-based object-oriented programming.

### Defining Classes
```javascript
class Person {
    // Constructor
    constructor(name, age) {
        this.name = name;
        this.age = age;
    }

    // Instance method
    greet() {
        return `Hello, I'm ${this.name}`;
    }

    // Getter
    get birthYear() {
        return new Date().getFullYear() - this.age;
    }

    // Setter
    set birthYear(year) {
        this.age = new Date().getFullYear() - year;
    }

    // Static method
    static createAnonymous() {
        return new Person("Anonymous", 0);
    }
}

// Inheritance
class Student extends Person {
    constructor(name, age, grade) {
        super(name, age); // Call parent constructor
        this.grade = grade;
    }

    // Override method
    greet() {
        return `${super.greet()} and I'm in grade ${this.grade}`;
    }

    // Additional method
    study() {
        return `${this.name} is studying`;
    }
}
```

### Using Classes
```javascript
// Create instances
let alice = new Person("Alice", 25);
let bob = new Student("Bob", 16, 11);

// Call methods
console.log(alice.greet());     // "Hello, I'm Alice"
console.log(bob.greet());       // "Hello, I'm Bob and I'm in grade 11"
console.log(bob.study());       // "Bob is studying"

// Use getters/setters
console.log(alice.birthYear);   // e.g., 1998
alice.birthYear = 1995;
console.log(alice.age);         // 28

// Static method
let anonymous = Person.createAnonymous();
console.log(anonymous.name);    // "Anonymous"
```

## Asynchronous JavaScript

JavaScript handles asynchronous operations using callbacks, promises, and async/await.

### Callbacks
```javascript
function fetchData(callback) {
    setTimeout(() => {
        callback("Data received");
    }, 1000);
}

fetchData((data) => {
    console.log(data); // "Data received" (after 1 second)
});
```

### Promises
```javascript
function fetchData() {
    return new Promise((resolve, reject) => {
        setTimeout(() => {
            resolve("Data received");
        }, 1000);
    });
}

fetchData()
    .then(data => {
        console.log(data); // "Data received"
        return data.toUpperCase();
    })
    .then(upperData => {
        console.log(upperData); // "DATA RECEIVED"
    })
    .catch(error => {
        console.error(error);
    });
```

### Async/Await (ES8+)
```javascript
async function getData() {
    try {
        let data = await fetchData();
        console.log(data); // "Data received"

        let processedData = await processData(data);
        console.log(processedData);
    } catch (error) {
        console.error(error);
    }
}

getData();
```

## DOM Manipulation (Browser)

JavaScript interacts with HTML documents through the DOM.

### Selecting Elements
```javascript
// By ID
let header = document.getElementById('main-header');

// By class name
let buttons = document.getElementsByClassName('btn');

// By tag name
let paragraphs = document.getElementsByTagName('p');

// By CSS selector
let firstButton = document.querySelector('.btn');
let allButtons = document.querySelectorAll('.btn');
```

### Modifying Elements
```javascript
let element = document.getElementById('myElement');

// Change content
element.textContent = "New text content";
element.innerHTML = "<strong>Bold text</strong>";

// Change styles
element.style.color = "red";
element.style.fontSize = "20px";

// Add/remove classes
element.classList.add('highlight');
element.classList.remove('hidden');
element.classList.toggle('active');
```

### Event Handling
```javascript
let button = document.getElementById('myButton');

// Add event listener
button.addEventListener('click', function(event) {
    console.log('Button clicked!');
    event.preventDefault(); // Prevent default action
});

// Event object properties
button.addEventListener('click', function(event) {
    console.log('Mouse position:', event.clientX, event.clientY);
    console.log('Target element:', event.target);
});

// Remove event listener
function handleClick() {
    console.log('Clicked!');
}

button.addEventListener('click', handleClick);
button.removeEventListener('click', handleClick);
```

## Modules (ES6+)

JavaScript supports modular code organization.

### Exporting
```javascript
// math.js
export function add(a, b) {
    return a + b;
}

export function multiply(a, b) {
    return a * b;
}

export const PI = 3.14159;

// Default export
export default function square(x) {
    return x * x;
}
```

### Importing
```javascript
// main.js
import { add, multiply, PI } from './math.js';
import square from './math.js'; // Default import

console.log(add(5, 3));      // 8
console.log(multiply(4, 2)); // 8
console.log(PI);             // 3.14159
console.log(square(5));      // 25
```

## Error Handling

JavaScript uses try-catch blocks for error handling.

```javascript
try {
    // Code that might throw an error
    let result = riskyOperation();
    console.log(result);
} catch (error) {
    console.error('An error occurred:', error.message);
} finally {
    // Always executes
    console.log('Cleanup code here');
}

// Throwing custom errors
function divide(a, b) {
    if (b === 0) {
        throw new Error('Division by zero is not allowed');
    }
    return a / b;
}

try {
    let result = divide(10, 0);
} catch (error) {
    console.error(error.message); // "Division by zero is not allowed"
}
```

## Modern JavaScript Features

### Destructuring
```javascript
// Array destructuring
let [first, second, ...rest] = [1, 2, 3, 4, 5];
console.log(first, second, rest); // 1, 2, [3, 4, 5]

// Object destructuring
let { name, age, city } = { name: "Alice", age: 25, city: "NYC" };
console.log(name, age, city); // "Alice", 25, "NYC"

// Default values
let { name = "Anonymous", age = 18 } = {};
console.log(name, age); // "Anonymous", 18
```

### Spread Operator
```javascript
// Array spreading
let arr1 = [1, 2, 3];
let arr2 = [4, 5, 6];
let combined = [...arr1, ...arr2]; // [1, 2, 3, 4, 5, 6]

// Object spreading
let person = { name: "Alice", age: 25 };
let updatedPerson = { ...person, city: "NYC" };

// Function arguments
function sum(a, b, c) {
    return a + b + c;
}
let numbers = [1, 2, 3];
console.log(sum(...numbers)); // 6
```

### Optional Chaining
```javascript
let user = {
    name: "Alice",
    address: {
        city: "NYC"
    }
};

// Safe property access
console.log(user.address?.city);    // "NYC"
console.log(user.contact?.phone);   // undefined (no error)

// Safe method calls
console.log(user.greet?.());        // undefined (no error if method doesn't exist)
```

### Nullish Coalescing
```javascript
// Returns right side only if left side is null or undefined
let name = null;
let displayName = name ?? "Anonymous"; // "Anonymous"

let age = 0;
let displayAge = age ?? 18; // 0 (not null/undefined, so left side used)
```

## Next Steps

You've covered the essentials of JavaScript! Here are some paths forward:

1. **Web Development**: Learn HTML, CSS, and frameworks like React
2. **Backend Development**: Explore Node.js and Express
3. **Full-Stack**: Combine frontend and backend skills
4. **Mobile Development**: Learn React Native
5. **Advanced Topics**: Closures, prototypes, design patterns

Practice regularly, build projects, and don't be afraid to experiment. JavaScript is a powerful language with a bright future!