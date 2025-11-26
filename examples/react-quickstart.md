# React Quickstart Guide

Get started with React, the popular JavaScript library for building user interfaces. This course covers the essentials of React development.

## What is React?

React is a JavaScript library for building user interfaces, particularly web applications. Created by Facebook (now Meta) and open-sourced in 2013, React has become one of the most popular frontend frameworks.

### Key Features
- **Component-Based**: Build encapsulated components that manage their own state
- **Declarative**: Describe what you want, React handles the DOM updates
- **Learn Once, Write Anywhere**: Use React for web, mobile (React Native), and more
- **Virtual DOM**: Efficiently updates only what changed
- **Unidirectional Data Flow**: Predictable data flow with props and state

## Setting Up React

### Create React App (Recommended for beginners)
```bash
# Install Node.js first (from nodejs.org)

# Create a new React app
npx create-react-app my-app

# Navigate to the project
cd my-app

# Start the development server
npm start
```

### Manual Setup (Advanced)
```bash
# Create package.json
npm init -y

# Install React and ReactDOM
npm install react react-dom

# Install development dependencies
npm install --save-dev webpack webpack-cli babel-loader @babel/core @babel/preset-react html-webpack-plugin webpack-dev-server
```

## JSX: JavaScript + XML

JSX is a syntax extension for JavaScript that allows you to write HTML-like code in your JavaScript files.

### Basic JSX
```jsx
// Without JSX
const element = React.createElement('h1', null, 'Hello, World!');

// With JSX
const element = <h1>Hello, World!</h1>;
```

### Embedding Expressions
```jsx
const name = 'Alice';
const element = <h1>Hello, {name}!</h1>;

// Expressions work too
const element = <h1>2 + 2 = {2 + 2}</h1>;

// Functions
function formatName(user) {
  return user.firstName + ' ' + user.lastName;
}

const user = { firstName: 'John', lastName: 'Doe' };
const element = <h1>Hello, {formatName(user)}!</h1>;
```

### JSX Attributes
```jsx
// String literals
const element = <div tabIndex="0"></div>;

// Curly braces for expressions
const element = <img src={user.avatarUrl} alt={user.name} />;

// Spread attributes
const props = { firstName: 'John', lastName: 'Doe' };
const element = <Greeting {...props} />;
```

### JSX and Children
```jsx
// Single child
const element = <div>Hello!</div>;

// Multiple children
const element = (
  <div>
    <h1>Hello!</h1>
    <p>Welcome to React</p>
  </div>
);

// Expressions as children
const items = ['Apple', 'Banana', 'Orange'];
const element = (
  <ul>
    {items.map(item => <li key={item}>{item}</li>)}
  </ul>
);
```

## Components

Components are the building blocks of React applications. They let you split the UI into independent, reusable pieces.

### Function Components
```jsx
function Welcome(props) {
  return <h1>Hello, {props.name}</h1>;
}

// Arrow function syntax
const Welcome = (props) => {
  return <h1>Hello, {props.name}</h1>;
};

// Implicit return (single expression)
const Welcome = (props) => <h1>Hello, {props.name}</h1>;
```

### Class Components
```jsx
class Welcome extends React.Component {
  render() {
    return <h1>Hello, {this.props.name}</h1>;
  }
}
```

### Component Composition
```jsx
function App() {
  return (
    <div>
      <Welcome name="Alice" />
      <Welcome name="Bob" />
      <Welcome name="Charlie" />
    </div>
  );
}
```

## Props

Props (properties) are how you pass data from parent components to child components.

### Using Props
```jsx
function Welcome(props) {
  return <h1>Hello, {props.name}!</h1>;
}

// Usage
<Welcome name="Alice" />

// Multiple props
function UserCard(props) {
  return (
    <div>
      <h2>{props.name}</h2>
      <p>Age: {props.age}</p>
      <p>Email: {props.email}</p>
    </div>
  );
}

// Usage
<UserCard name="John Doe" age={30} email="john@example.com" />
```

### Default Props
```jsx
function Greeting(props) {
  return <h1>Hello, {props.name}!</h1>;
}

Greeting.defaultProps = {
  name: 'Guest'
};

// Or with default parameters
function Greeting({ name = 'Guest' }) {
  return <h1>Hello, {name}!</h1>;
}
```

### Prop Types (Type Checking)
```jsx
import PropTypes from 'prop-types';

function UserCard(props) {
  return (
    <div>
      <h2>{props.name}</h2>
      <p>{props.description}</p>
    </div>
  );
}

UserCard.propTypes = {
  name: PropTypes.string.isRequired,
  description: PropTypes.string,
  age: PropTypes.number,
  isActive: PropTypes.bool,
  hobbies: PropTypes.arrayOf(PropTypes.string),
  address: PropTypes.shape({
    street: PropTypes.string,
    city: PropTypes.string,
    zipCode: PropTypes.string
  })
};
```

## State

State allows components to manage and update their own data.

### useState Hook
```jsx
import React, { useState } from 'react';

function Counter() {
  const [count, setCount] = useState(0);

  return (
    <div>
      <p>You clicked {count} times</p>
      <button onClick={() => setCount(count + 1)}>
        Click me
      </button>
    </div>
  );
}
```

### Multiple State Variables
```jsx
function UserForm() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [age, setAge] = useState(18);

  return (
    <form>
      <input
        value={name}
        onChange={e => setName(e.target.value)}
        placeholder="Name"
      />
      <input
        value={email}
        onChange={e => setEmail(e.target.value)}
        placeholder="Email"
      />
      <input
        type="number"
        value={age}
        onChange={e => setAge(parseInt(e.target.value))}
        placeholder="Age"
      />
    </form>
  );
}
```

### State Updates
```jsx
function Counter() {
  const [count, setCount] = useState(0);

  // Correct: Use functional updates for current state
  const increment = () => {
    setCount(prevCount => prevCount + 1);
  };

  // Incorrect: This might not work as expected
  const badIncrement = () => {
    setCount(count + 1); // Uses stale closure value
  };

  return (
    <div>
      <p>Count: {count}</p>
      <button onClick={increment}>Increment</button>
    </div>
  );
}
```

### Complex State
```jsx
function TodoApp() {
  const [todos, setTodos] = useState([]);
  const [inputValue, setInputValue] = useState('');

  const addTodo = () => {
    if (inputValue.trim()) {
      setTodos([...todos, { id: Date.now(), text: inputValue, completed: false }]);
      setInputValue('');
    }
  };

  const toggleTodo = (id) => {
    setTodos(todos.map(todo =>
      todo.id === id ? { ...todo, completed: !todo.completed } : todo
    ));
  };

  const deleteTodo = (id) => {
    setTodos(todos.filter(todo => todo.id !== id));
  };

  return (
    <div>
      <input
        value={inputValue}
        onChange={e => setInputValue(e.target.value)}
        placeholder="Add a todo"
      />
      <button onClick={addTodo}>Add</button>

      <ul>
        {todos.map(todo => (
          <li key={todo.id}>
            <span
              style={{ textDecoration: todo.completed ? 'line-through' : 'none' }}
              onClick={() => toggleTodo(todo.id)}
            >
              {todo.text}
            </span>
            <button onClick={() => deleteTodo(todo.id)}>Delete</button>
          </li>
        ))}
      </ul>
    </div>
  );
}
```

## Effects

useEffect allows you to perform side effects in function components.

### Basic useEffect
```jsx
import React, { useState, useEffect } from 'react';

function Example() {
  const [count, setCount] = useState(0);

  // Runs after every render
  useEffect(() => {
    document.title = `You clicked ${count} times`;
  });

  return (
    <div>
      <p>You clicked {count} times</p>
      <button onClick={() => setCount(count + 1)}>
        Click me
      </button>
    </div>
  );
}
```

### Conditional Effects
```jsx
function UserProfile({ userId }) {
  const [user, setUser] = useState(null);

  useEffect(() => {
    // Only fetch when userId changes
    fetchUser(userId).then(setUser);
  }, [userId]); // Dependency array

  if (!user) return <div>Loading...</div>;

  return <div>{user.name}</div>;
}
```

### Cleanup
```jsx
function Timer() {
  const [seconds, setSeconds] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => {
      setSeconds(seconds => seconds + 1);
    }, 1000);

    // Cleanup function
    return () => clearInterval(interval);
  }, []); // Empty dependency array = run once

  return <div>Seconds: {seconds}</div>;
}
```

## Event Handling

React provides a consistent way to handle events.

### Basic Event Handling
```jsx
function Button() {
  const handleClick = () => {
    console.log('Button clicked!');
  };

  return <button onClick={handleClick}>Click me</button>;
}
```

### Event Object
```jsx
function Form() {
  const handleSubmit = (event) => {
    event.preventDefault(); // Prevent default form submission
    console.log('Form submitted');
  };

  const handleChange = (event) => {
    console.log('Input value:', event.target.value);
  };

  return (
    <form onSubmit={handleSubmit}>
      <input onChange={handleChange} />
      <button type="submit">Submit</button>
    </form>
  );
}
```

### Passing Arguments to Event Handlers
```jsx
function List() {
  const items = ['Apple', 'Banana', 'Orange'];

  const handleClick = (item) => {
    console.log('Clicked:', item);
  };

  return (
    <ul>
      {items.map(item => (
        <li key={item} onClick={() => handleClick(item)}>
          {item}
        </li>
      ))}
    </ul>
  );
}
```

## Conditional Rendering

React allows you to render different content based on conditions.

### If-Else
```jsx
function Greeting(props) {
  if (props.isLoggedIn) {
    return <h1>Welcome back!</h1>;
  } else {
    return <h1>Please sign in.</h1>;
  }
}
```

### Ternary Operator
```jsx
function Greeting(props) {
  return (
    <h1>
      {props.isLoggedIn ? 'Welcome back!' : 'Please sign in.'}
    </h1>
  );
}
```

### Logical AND
```jsx
function Mailbox(props) {
  return (
    <div>
      <h1>Hello!</h1>
      {props.unreadMessages.length > 0 && (
        <h2>You have {props.unreadMessages.length} unread messages.</h2>
      )}
    </div>
  );
}
```

### Preventing Rendering
```jsx
function WarningBanner(props) {
  if (!props.warn) {
    return null; // Don't render anything
  }

  return <div className="warning">Warning!</div>;
}
```

## Lists and Keys

When rendering lists, each item needs a unique key.

### Basic Lists
```jsx
function NumberList(props) {
  const numbers = props.numbers;

  const listItems = numbers.map(number => (
    <li key={number.toString()}>
      {number}
    </li>
  ));

  return <ul>{listItems}</ul>;
}
```

### Keys
```jsx
// Good: Use stable IDs
const todoItems = todos.map(todo => (
  <li key={todo.id}>
    {todo.text}
  </li>
));

// OK: Use array index (only if items never reorder)
const items = items.map((item, index) => (
  <li key={index}>
    {item}
  </li>
));

// Bad: Random keys cause unnecessary re-renders
const items = items.map(item => (
  <li key={Math.random()}>
    {item}
  </li>
));
```

## Forms

React provides controlled components for form handling.

### Controlled Components
```jsx
function NameForm() {
  const [value, setValue] = useState('');

  const handleChange = (event) => {
    setValue(event.target.value);
  };

  const handleSubmit = (event) => {
    event.preventDefault();
    alert('A name was submitted: ' + value);
  };

  return (
    <form onSubmit={handleSubmit}>
      <label>
        Name:
        <input type="text" value={value} onChange={handleChange} />
      </label>
      <input type="submit" value="Submit" />
    </form>
  );
}
```

### Multiple Inputs
```jsx
function Reservation() {
  const [reservation, setReservation] = useState({
    isGoing: true,
    numberOfGuests: 2
  });

  const handleInputChange = (event) => {
    const target = event.target;
    const value = target.type === 'checkbox' ? target.checked : target.value;
    const name = target.name;

    setReservation({
      ...reservation,
      [name]: value
    });
  };

  return (
    <form>
      <label>
        Is going:
        <input
          name="isGoing"
          type="checkbox"
          checked={reservation.isGoing}
          onChange={handleInputChange}
        />
      </label>
      <br />
      <label>
        Number of guests:
        <input
          name="numberOfGuests"
          type="number"
          value={reservation.numberOfGuests}
          onChange={handleInputChange}
        />
      </label>
    </form>
  );
}
```

## Lifting State Up

When multiple components need to share state, lift it up to their common parent.

```jsx
function BoilingVerdict(props) {
  if (props.celsius >= 100) {
    return <p>The water would boil.</p>;
  }
  return <p>The water would not boil.</p>;
}

function Calculator() {
  const [temperature, setTemperature] = useState('');

  const handleChange = (event) => {
    setTemperature(event.target.value);
  };

  return (
    <fieldset>
      <legend>Enter temperature in Celsius:</legend>
      <input value={temperature} onChange={handleChange} />
      <BoilingVerdict celsius={parseFloat(temperature)} />
    </fieldset>
  );
}
```

## Context

Context provides a way to pass data through the component tree without having to pass props down manually at every level.

### Creating Context
```jsx
const ThemeContext = React.createContext('light');

function App() {
  return (
    <ThemeContext.Provider value="dark">
      <Toolbar />
    </ThemeContext.Provider>
  );
}
```

### Consuming Context
```jsx
function ThemedButton() {
  const theme = useContext(ThemeContext);

  return (
    <button style={{ background: theme === 'dark' ? '#333' : '#fff' }}>
      Themed Button
    </button>
  );
}
```

## Hooks

Hooks are functions that let you "hook into" React state and lifecycle features from function components.

### useState
```jsx
const [state, setState] = useState(initialState);
```

### useEffect
```jsx
useEffect(() => {
  // Side effect code
  return () => {
    // Cleanup code
  };
}, [dependencies]);
```

### useContext
```jsx
const value = useContext(MyContext);
```

### useReducer (for complex state)
```jsx
const [state, dispatch] = useReducer(reducer, initialState);
```

### Custom Hooks
```jsx
function useLocalStorage(key, initialValue) {
  const [storedValue, setStoredValue] = useState(() => {
    try {
      const item = window.localStorage.getItem(key);
      return item ? JSON.parse(item) : initialValue;
    } catch (error) {
      return initialValue;
    }
  });

  const setValue = value => {
    try {
      setStoredValue(value);
      window.localStorage.setItem(key, JSON.stringify(value));
    } catch (error) {
      console.error(error);
    }
  };

  return [storedValue, setValue];
}
```

## Best Practices

### Component Structure
- Keep components small and focused
- Use descriptive names
- Extract reusable logic into custom hooks
- Prefer function components over class components

### State Management
- Keep state as local as possible
- Lift state up when needed
- Use useReducer for complex state logic
- Consider Context API or Redux for global state

### Performance
- Use React.memo for expensive components
- Use useMemo for expensive calculations
- Use useCallback for event handlers
- Avoid unnecessary re-renders

### Code Organization
```
src/
├── components/
│   ├── common/
│   ├── layout/
│   └── features/
├── hooks/
├── contexts/
├── services/
├── utils/
└── types/
```

## Next Steps

You've learned the fundamentals of React! Here's what to explore next:

1. **Advanced Hooks**: useMemo, useCallback, useRef, custom hooks
2. **Routing**: React Router for multi-page apps
3. **State Management**: Redux, Zustand, or Context API
4. **Testing**: Jest and React Testing Library
5. **Styling**: CSS Modules, styled-components, or Tailwind CSS
6. **TypeScript**: Add type safety to your React apps
7. **Next.js**: Full-stack React framework
8. **React Native**: Build mobile apps with React

Practice by building small projects and gradually add complexity. The React ecosystem is vast and constantly evolving!