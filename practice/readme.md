Here's a quick Go (Golang) cheatsheet to help you master the language with code examples:

## Basics

### 1. Hello World
```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

### 2. Variables and Constants

#### Variables
```go
// Declaring a single variable
var name string = "John"

// Type inferred
age := 30

// Multiple variables
var (
    x int
    y int = 10
    z = 20
)
```

#### Constants
```go
const Pi = 3.14
const (
    Hello = "Hello"
    World = "World"
)
```

### 3. Data Types
```go
var a bool = true
var b int = 42
var c float64 = 3.14
var d string = "Hello"
```

### 4. Arrays and Slices

#### Arrays
```go
var arr [3]int // Array of three integers
arr[0] = 1
arr[1] = 2
arr[2] = 3

// Declaration & Initialization
numbers := [3]int{1, 2, 3}
```

#### Slices
```go
slice := []int{1, 2, 3}
slice = append(slice, 4)
```

### 5. Maps
```go
ages := map[string]int{
    "Alice": 25,
    "Bob": 30,
}

// Adding an element
ages["Charlie"] = 35

// Accessing an element
age := ages["Alice"]

// Deleting an element
delete(ages, "Bob")
```

### 6. Functions
```go
func add(x int, y int) int {
    return x + y
}

// Multiple return values
func swap(x, y string) (string, string) {
    return y, x
}

func main() {
    fmt.Println(add(1, 2))

    a, b := swap("hello", "world")
    fmt.Println(a, b)
}
```

### 7. Control Structures

#### If-Else
```go
if x > 0 {
    fmt.Println("x is positive")
} else {
    fmt.Println("x is not positive")
}
```

#### Switch
```go
switch day {
case "Monday":
    fmt.Println("Start of the work week")
case "Friday":
    fmt.Println("Almost weekend")
default:
    fmt.Println("Midweek")
}
```

#### For Loop
```go
// Standard for loop
for i := 0; i < 5; i++ {
    fmt.Println(i)
}

// For loop as a while loop
sum := 1
for sum < 1000 {
    sum += sum
}
fmt.Println(sum)

// Infinite loop
for {
    fmt.Println("loop")
}
```

#### Range
```go
// Range with slice
nums := []int{2, 3, 4}
for i, num := range nums {
    fmt.Printf("Index %d, Value %d\n", i, num)
}

// Range with map
kvs := map[string]string{"a": "apple", "b": "banana"}
for k, v := range kvs {
    fmt.Printf("%s -> %s\n", k, v)
}
```

### 8. Structs
```go
type Person struct {
    Name string
    Age  int
}

func main() {
    // Creating a struct
    p := Person{"Alice", 30}

    // Accessing fields
    fmt.Println(p.Name, p.Age)

    // Anonymous struct
    point := struct {
        x, y int
    }{10, 20}
    fmt.Println(point)
}
```

### 9. Methods
```go
type Circle struct {
    Radius float64
}

// Method on Circle struct
func (c Circle) Area() float64 {
    return 3.14 * c.Radius * c.Radius
}

func main() {
    c := Circle{10}
    fmt.Println("Area of Circle:", c.Area())
}
```

### 10. Interfaces
```go
type Shape interface {
    Area() float64
}

type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func printArea(s Shape) {
    fmt.Println("Area:", s.Area())
}

func main() {
    r := Rectangle{10, 20}
    printArea(r)
}
```

### 11. Goroutines and Channels

#### Goroutines
```go
func say(s string) {
    for i := 0; i < 5; i++ {
        fmt.Println(s)
        time.Sleep(100 * time.Millisecond)
    }
}

func main() {
    go say("world")
    say("hello")
}
```

#### Channels
```go
func sum(s []int, c chan int) {
    sum := 0
    for _, v := range s {
        sum += v
    }
    c <- sum // send sum to c
}

func main() {
    s := []int{7, 2, 8, -9, 4, 0}

    c := make(chan int)
    go sum(s[:len(s)/2], c)
    go sum(s[len(s)/2:], c)
    x, y := <-c, <-c // receive from c

    fmt.Println(x, y, x+y)
}
```

### 12. Error Handling
```go
func division(a, b int) (int, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}

func main() {
    if result, err := division(4, 2); err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Result:", result)
    }
}
```

### 13. Packages and Modules
#### Basic Package Usage
```go
// file: mypackage/mypackage.go
package mypackage

import "fmt"

func SayHello(name string) {
    fmt.Println("Hello", name)
}
```

#### Using a Custom Package
```go
// file: main.go
package main

import "mypackage"

func main() {
    mypackage.SayHello("GoLang")
}
```

#### Creating and Using Modules
1. **Initialize a new module:**
   ```sh
   go mod init example.com/mymodule
   ```

2. **Add dependencies:**
   - Add import statements in your code, and run the build command to add them to `go.mod` automatically.
   - Alternatively, use `go get` to manually add dependencies.

3. **Format and tidy up module information:**
   ```sh
   go mod tidy
   ```

This cheatsheet should help you quickly catch up with Go's syntax and capabilities, enabling you to write more efficient and effective Go code.