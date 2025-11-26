# Introduction to Go Programming

This comprehensive course will teach you the fundamentals of Go programming language, from basic syntax to advanced concepts.

## What is Go?

Go is a statically typed, compiled programming language designed at Google. It's known for its simplicity, efficiency, and excellent support for concurrent programming.

### Key Features

- **Simple syntax**: Clean and readable code
- **Fast compilation**: Quick build times
- **Garbage collection**: Automatic memory management
- **Concurrency**: Built-in support for goroutines

## Setting Up Your Environment

Before we start coding, let's set up your Go development environment.

### Installing Go

1. Download the Go installer from golang.org
2. Run the installer
3. Verify installation: `go version`

### Your First Go Program

Let's create a simple "Hello, World!" program.

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

Save this as `main.go` and run it with `go run main.go`.

## Basic Syntax

Go has a clean and straightforward syntax. Let's explore the basic building blocks.

### Variables

```go
// Variable declarations
var name string = "Go"
age := 10  // Short declaration

// Constants
const Pi = 3.14159
```

### Functions

```go
func add(a int, b int) int {
    return a + b
}

// Multiple return values
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}
```

## Control Structures

Go provides familiar control structures with some unique features.

### If Statements

```go
if x > 10 {
    fmt.Println("x is greater than 10")
} else if x < 5 {
    fmt.Println("x is less than 5")
} else {
    fmt.Println("x is between 5 and 10")
}
```

### Loops

Go has only one type of loop: `for`.

```go
// Basic for loop
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// While-like loop
j := 0
for j < 10 {
    fmt.Println(j)
    j++
}

// Infinite loop
for {
    // Do something forever
}
```

## Data Structures

Go provides several built-in data structures.

### Arrays and Slices

```go
// Array
var arr [5]int

// Slice (dynamic array)
numbers := []int{1, 2, 3, 4, 5}

// Append to slice
numbers = append(numbers, 6)
```

### Maps

```go
// Create a map
person := map[string]string{
    "name": "John",
    "age":  "30",
}

// Access map
name := person["name"]

// Add to map
person["city"] = "New York"
```

## Structs

Structs are Go's way of creating custom data types.

```go
type Person struct {
    Name string
    Age  int
    City string
}

// Create struct instance
john := Person{
    Name: "John",
    Age:  30,
    City: "New York",
}

// Access fields
fmt.Println(john.Name)
```

## Methods

Go supports methods on custom types.

```go
func (p Person) greet() string {
    return "Hello, my name is " + p.Name
}

// Call method
message := john.greet()
```

## Concurrency

One of Go's strongest features is its built-in concurrency support.

### Goroutines

Goroutines are lightweight threads.

```go
func sayHello() {
    fmt.Println("Hello from goroutine!")
}

func main() {
    go sayHello()  // Start goroutine
    time.Sleep(time.Second)  // Wait for goroutine
}
```

### Channels

Channels are used for communication between goroutines.

```go
func worker(id int, jobs <-chan int, results chan<- int) {
    for j := range jobs {
        results <- j * 2  // Send result
    }
}

func main() {
    jobs := make(chan int, 100)
    results := make(chan int, 100)

    // Start workers
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }

    // Send jobs
    for j := 1; j <= 5; j++ {
        jobs <- j
    }
    close(jobs)

    // Collect results
    for a := 1; a <= 5; a++ {
        <-results
    }
}
```

## Error Handling

Go uses explicit error handling instead of exceptions.

```go
func readFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    // Read file contents
    return nil
}

func main() {
    err := readFile("example.txt")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("File read successfully")
}
```

## Packages and Modules

Go organizes code into packages and modules.

### Creating a Package

Create a file `math.go`:

```go
package math

func Add(a, b int) int {
    return a + b
}
```

### Using the Package

```go
package main

import (
    "fmt"
    "yourmodule/math"
)

func main() {
    result := math.Add(5, 3)
    fmt.Println(result)  // Prints: 8
}
```

## Testing

Go has built-in support for testing.

### Writing Tests

Create a file `math_test.go`:

```go
package math

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    expected := 5

    if result != expected {
        t.Errorf("Add(2, 3) = %d; want %d", result, expected)
    }
}
```

Run tests with `go test`.

## Next Steps

Congratulations! You've learned the basics of Go programming. Here are some next steps:

1. **Practice**: Write small programs to reinforce your learning
2. **Read the docs**: Check out the official Go documentation
3. **Join the community**: Participate in Go forums and meetups
4. **Build projects**: Create real applications using Go

Remember, the best way to learn programming is by doing. Start building!