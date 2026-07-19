# Learn Go: The Complete Course

> Adapted from [Karan Pratap Singh's "Learn Go: The Complete Course"](https://www.karanpratapsingh.com/courses/go) (originally published May 2022), reformatted for readability and updated for modern Go (this project targets **Go 1.24+**, see [`coding-standards.md`](coding-standards.md)). Wherever the language has changed since the original course was written, that's called out explicitly in a **📌 Updated** note rather than silently — the surrounding teaching content is still accurate and worth learning in the original order.
>
> This is a personal learning reference for working through Go while building this project — not original project documentation. Keep it around, add your own notes as you go.

---

## Table of Contents

1. [What is Go?](#what-is-go)
2. [Installation and Setup](#installation-and-setup)
3. [Hello World](#hello-world)
4. [Variables and Data Types](#variables-and-data-types)
5. [String Formatting](#string-formatting)
6. [Flow Control](#flow-control)
7. [Functions](#functions)
8. [Modules](#modules)
9. [Packages](#packages)
10. [Workspaces](#workspaces)
11. [Useful Commands](#useful-commands)
12. [Build](#build)
13. [Pointers](#pointers)
14. [Structs](#structs)
15. [Methods](#methods)
16. [Arrays and Slices](#arrays-and-slices)
17. [Maps](#maps)
18. [Interfaces](#interfaces)
19. [Errors](#errors)
20. [Panic and Recover](#panic-and-recover)
21. [Testing](#testing)
22. [Generics](#generics)
23. [Concurrency](#concurrency)
24. [Goroutines](#goroutines)
25. [Channels](#channels)
26. [Select](#select)
27. [Sync Package](#sync-package)
28. [Advanced Concurrency Patterns](#advanced-concurrency-patterns)
29. [Context](#context)
30. [Newer Standard Library Additions Worth Knowing](#newer-standard-library-additions-worth-knowing)
31. [Next Steps](#next-steps)

---

## What is Go?

Go (also known as Golang) is a programming language developed at Google in 2007 and open-sourced in 2009.

It focuses on simplicity, reliability, and efficiency. It was designed to combine the efficacy, speed, and safety of a statically typed, compiled language with the ease of programming of a dynamic language.

In a way, it combines the best parts of Python and C++, so you can build reliable systems that take advantage of multi-core processors without fighting the language to do it.

### Why Learn Go?

1. **Easy to learn.** Go has a small surface area and a supportive, active community. As a multipurpose language, you can use it for backend development, cloud computing, and increasingly, data tooling.
2. **Fast and reliable.** Highly suitable for distributed systems — Kubernetes and Docker are both written in Go.
3. **Simple yet powerful.** Go has just 25 keywords, which keeps it easy to read, write, and maintain. Don't mistake that simplicity for a lack of power — the language has several deep features you'll learn throughout this course.
4. **Career opportunities.** Go is growing fast and is adopted by companies of every size.

---

## Installation and Setup

### Download and Install

Get the current release from the [official downloads page](https://go.dev/dl/) — always prefer that over a hardcoded version number, since Go ships new releases roughly every six months.

**macOS** — open the downloaded package and follow the prompts. The installer places the distribution at `/usr/local/go` and adds `/usr/local/go/bin` to your `PATH`. Restart any open terminal sessions afterward.

```bash
go version
```

**Linux** — remove any previous installation, then extract the archive into `/usr/local` (do **not** extract into an existing `/usr/local/go` tree — that produces a broken install):

```bash
rm -rf /usr/local/go
tar -C /usr/local -xzf go<version>.linux-<arch>.tar.gz   # use the exact filename you downloaded
```

Add Go to your `PATH` (in `$HOME/.profile` or `/etc/profile` for a system-wide install):

```bash
export PATH=$PATH:/usr/local/go/bin
```

Changes to a profile file don't apply until your next login — run `source $HOME/.profile` (or the shell command directly) to apply them immediately.

```bash
go version
```

**Windows** — open the downloaded MSI and follow the prompts. By default it installs to `Program Files` (or `Program Files (x86)`); close and reopen any command prompts afterward so the environment changes take effect.

```cmd
go version
```

### Editor

Any editor works, but [VS Code](https://code.visualstudio.com) with the official [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go) is the most common setup — it wires up `gopls` (the Go language server), inline diagnostics, and `go test`/`go build` integration out of the box.

---

## Hello World

Start by initializing a module — a module is a collection of Go packages (more on this in [Modules](#modules)):

```bash
go mod init example
```

Then create `main.go`:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello World!")
}
```

`fmt` is part of the Go standard library — the set of core packages that ship with the language itself.

### Structure of a Go Program

Every Go source file starts with a package declaration:

```go
package main
```

Then any imports:

```go
import "fmt"
```

Then the code itself. `main()` is the entry point for an executable program, the same role it plays in C, Java, or C#:

```go
func main() {
    // ...
}
```

Run it with:

```bash
go run main.go
```

```
Hello World!
```

---

## Variables and Data Types

### Variables

Declaration without initialization:

```go
var foo string
```

Declaration with initialization:

```go
var foo string = "Go is awesome"
```

Multiple declarations:

```go
var foo, bar string = "Hello", "World"

// or, grouped:
var (
    foo string = "Hello"
    bar string = "World"
)
```

Type omitted, inferred from the value:

```go
var foo = "What's my type?"
```

**Shorthand declaration** — omits `var`, the type is always inferred, and `:=` does declaration + assignment in one step. This is how you'll see most variables declared in real Go code:

```go
foo := "Shorthand!"
```

> Shorthand (`:=`) only works inside function bodies — not at package level.

### Constants

```go
const constant = "This is a constant"
```

Only constants can be assigned to other constants:

```go
const a = 10
const b = a // ✅ works

var a = 10
const b = a // ❌ a (variable of type int) is not constant
```

### Data Types

#### String

A string is a sequence of bytes, declared with double quotes or backticks (which can span multiple lines):

```go
var name string = "My name is Go"

var bio string = `I am statically typed.
I was designed at Google.`
```

#### Bool

```go
var value bool = false
var isItTrue bool = true
```

| Type | Syntax |
|---|---|
| Logical | `&&` `\|\|` `!` |
| Equality | `==` `!=` |

#### Numeric Types

Go has several built-in integer types of varying sizes, for both signed and unsigned integers. `int`/`uint` are platform-dependent — 32 bits wide on a 32-bit system, 64 bits on a 64-bit one.

```go
var i int = 404          // platform-dependent
var i8 int8 = 127        // -128 to 127
var i16 int16 = 32767    // -2^15 to 2^15-1
var i32 int32 = -2147483647 // -2^31 to 2^31-1
var i64 int64 = 9223372036854775807 // -2^63 to 2^63-1
```

```go
var ui uint = 404
var ui8 uint8 = 255       // 0 to 255
var ui16 uint16 = 65535   // 0 to 2^16
var ui32 uint32 = 2147483647 // 0 to 2^32
var ui64 uint64 = 9223372036854775807 // 0 to 2^64
var uiptr uintptr         // integer representation of a memory address
```

`uintptr` is rarely used directly — you generally don't need to worry about it. **Default to plain `int`** unless you have a specific reason to reach for a sized or unsigned type.

**Byte and rune** — aliases for `uint8` and `int32`:

```go
type byte = uint8
type rune = int32
```

A rune represents a Unicode code point:

```go
var b byte = 'a'
var r rune = '🍕'
```

**Floating point** — `float32` and `float64`, both IEEE-754. `float64` is the default:

```go
var f32 float32 = 1.7812 // IEEE-754 32-bit
var f64 float64 = 3.1415 // IEEE-754 64-bit
```

| Category | Syntax |
|---|---|
| Arithmetic | `+` `-` `*` `/` `%` |
| Comparison | `==` `!=` `<` `>` `<=` `>=` |
| Bitwise | `&` `\|` `^` `<<` `>>` |
| Increment/Decrement | `++` `--` |
| Assignment | `=` `+=` `-=` `*=` `/=` `%=` `<<=` `>>=` `&=` `\|=` `^=` |

**Complex** — `complex128` (real/imaginary as `float64`) and `complex64` (as `float32`):

```go
var c1 complex128 = complex(10, 1)
var c2 complex64 = 12 + 4i
```

### Zero Values

Any variable declared without an explicit initial value gets its **zero value** — this is different from most languages, where an unassigned variable is `null`/`undefined`:

```go
var i int
var f float64
var b bool
var s string

fmt.Printf("%v %v %v %q\n", i, f, b, s)
```

```
0 0 false ""
```

### Type Conversion

```go
i := 42
f := float64(i)
u := uint(f)

fmt.Printf("%T %T", f, u)
```

```
float64 uint
```

This is different from parsing a string into a number — conversion changes the representation of an already-typed value.

### Alias Types

Introduced in Go 1.9 — an alternate name for an existing type, usable interchangeably with the underlying type:

```go
package main

import "fmt"

type MyAlias = string

func main() {
    var str MyAlias = "I am an alias"
    fmt.Printf("%T - %s", str, str) // string - I am an alias
}
```

### Defined Types

Unlike alias types, defined types don't use an equals sign — and they create a genuinely distinct type, not interchangeable with the underlying one:

```go
package main

import "fmt"

type MyDefined string

func main() {
    var str MyDefined = "I am defined"
    fmt.Printf("%T - %s", str, str) // main.MyDefined - I am defined
}
```

```go
type MyAlias = string
type MyDefined string

func main() {
    var alias MyAlias
    var def MyDefined

    var copy1 string = alias // ✅ works
    var copy2 string = def   // ❌ cannot use def (MyDefined) as string value
}
```

---

## String Formatting

`fmt.Print` — no formatting, just concatenates:

```go
fmt.Print("What", "is", "your", "name?")
```

```
Whatisyourname?
```

`fmt.Println` — adds a newline and a space between arguments:

```go
fmt.Println("What", "is", "your", "name?")
```

```
What is your name?
```

`fmt.Printf` — "print formatter," substitutes values using **verbs**:

```go
name := "golang"
fmt.Printf("My name is %s", name)
```

```
My name is golang
```

`%s` and friends are annotation verbs controlling width, type, and precision — see the [fmt package docs](https://pkg.go.dev/fmt) for the full reference.

```go
percent := (7.0 / 9) * 100
fmt.Printf("%.2f %%", percent) // 2-decimal precision; %% escapes a literal percent sign
```

```
77.78 %
```

`Sprint`/`Sprintln`/`Sprintf` — same as the print family, but **return** the string instead of printing it:

```go
s := fmt.Sprintf("hex:%x bin:%b", 10, 10)
fmt.Println(s)
```

```
hex:a bin:1010
```

Multiline string literals:

```go
msg := `Hello from
multiline`
```

This is worth internalizing if you're coming from Python or JavaScript — `fmt`'s verb-based formatting feels unusual at first but is used extensively throughout idiomatic Go code.

---

## Flow Control

### If/Else

No parentheses required around the condition:

```go
func main() {
    x := 10

    if x > 5 {
        fmt.Println("x is gt 5")
    } else if x > 10 {
        fmt.Println("x is gt 10")
    } else {
        fmt.Println("else case")
    }
}
```

**Compact if** — a common pattern, scoping a variable to just the if/else:

```go
if x := 10; x > 5 {
    fmt.Println("x is gt 5")
}
```

### Switch

Only the first matching case runs — Go doesn't fall through to subsequent cases by default (the opposite of C-family languages):

```go
func main() {
    day := "monday"

    switch day {
    case "monday":
        fmt.Println("time to work!")
    case "friday":
        fmt.Println("let's party")
    default:
        fmt.Println("browse memes")
    }
}
```

Shorthand declaration:

```go
switch day := "monday"; day {
case "monday":
    fmt.Println("time to work!")
default:
    fmt.Println("browse memes")
}
```

`fallthrough` explicitly transfers control to the next case:

```go
switch day := "monday"; day {
case "monday":
    fmt.Println("time to work!")
    fallthrough
case "friday":
    fmt.Println("let's party")
default:
    fmt.Println("browse memes")
}
```

```
time to work!
let's party
```

A conditionless switch is shorthand for `switch true`:

```go
x := 10

switch {
case x > 5:
    fmt.Println("x is greater")
default:
    fmt.Println("x is not greater")
}
```

### Loops

Go has exactly one loop construct — `for` — but it's versatile enough to cover every case other languages use `while`/`do-while`/`foreach` for. No parentheses required.

**Basic for loop**, three semicolon-separated components — init (runs once before the first iteration), condition (checked before every iteration), post (runs at the end of every iteration):

```go
func main() {
    for i := 0; i < 10; i++ {
        fmt.Println(i)
    }
}
```

**Break and continue:**

```go
func main() {
    for i := 0; i < 10; i++ {
        if i < 2 {
            continue
        }

        fmt.Println(i)

        if i > 5 {
            break
        }
    }

    fmt.Println("We broke out!")
}
```

Init and post are optional — omit both to get `while`-loop behavior:

```go
func main() {
    i := 0

    for i < 10 {
        i += 1
    }
}
```

**Forever loop** — omit the condition entirely for an infinite loop:

```go
func main() {
    for {
        // do stuff here
    }
}
```

> **📌 Updated — Go 1.22 changed loop variable semantics.** Before 1.22, a `for`-loop variable was shared across all iterations — capturing it in a closure or goroutine (`for _, x := range xs { go func() { use(x) }() }`) was a well-known footgun, since every closure saw whatever `x` happened to be by the time it actually ran, usually the last value. As of Go 1.22, each iteration gets its **own** copy of the loop variable — that entire bug class is gone by default, not just discouraged by convention. If you're running an older Go version, you'd work around it by shadowing (`x := x`) inside the loop body; on 1.22+ that workaround is no longer necessary, though it's harmless if you see it in older code.
>
> **📌 Updated — Go 1.22 also added range-over-integers.** `for i := range 10 { ... }` iterates `i` from `0` to `9` — a cleaner way to write a counting loop than `for i := 0; i < 10; i++`.

---

## Functions

**Simple declaration:**

```go
func myFunction() {}
```

```go
func main() {
    myFunction("Hello")
}

func myFunction(p1 string) {
    fmt.Println(p1)
}
```

Consecutive parameters of the same type can share a type annotation:

```go
func myNextFunction(p1, p2 string) {}
```

**Returning a value:**

```go
func main() {
    s := myFunction("Hello")
    fmt.Println(s)
}

func myFunction(p1 string) string {
    msg := fmt.Sprintf("%s function", p1)
    return msg
}
```

**Multiple returns:**

```go
func main() {
    s, i := myFunction("Hello")
    fmt.Println(s, i)
}

func myFunction(p1 string) (string, int) {
    msg := fmt.Sprintf("%s function", p1)
    return msg, 10
}
```

**Named returns** — return values can be named and treated as their own variables, then returned via a bare `return` (a "naked return"):

```go
func myFunction(p1 string) (s string, i int) {
    s = fmt.Sprintf("%s function", p1)
    i = 10
    return
}
```

Useful, but use it carefully — naked returns can hurt readability in larger functions.

**Functions as values** — Go functions are first-class:

```go
func myFunction() {
    fn := func() {
        fmt.Println("inside fn")
    }

    fn()
}
```

Or anonymous and immediately invoked:

```go
func myFunction() {
    func() {
        fmt.Println("inside fn")
    }()
}
```

**Closures** — a function value that references variables from outside its own body. Closures are lexically scoped: a function can access whatever's in scope where it was *defined*:

```go
func myFunction() func(int) int {
    sum := 0

    return func(v int) int {
        sum += v
        return sum
    }
}
```

```go
add := myFunction()

add(5)
fmt.Println(add(10)) // 15 — sum is bound to the returned function
```

**Variadic functions** — take zero or more arguments via `...`:

```go
func main() {
    sum := add(1, 2, 3, 5)
    fmt.Println(sum)
}

func add(values ...int) int {
    sum := 0
    for _, v := range values {
        sum += v
    }
    return sum
}
```

(`fmt.Println` itself is variadic — that's how it accepts any number of arguments.)

### Init

`init` is a special lifecycle function that runs before `main`. Like `main`, it takes no arguments and returns nothing:

```go
package main

import "fmt"

func init() {
    fmt.Println("Before main!")
}

func main() {
    fmt.Println("Running main")
}
```

```
Before main!
Running main
```

Unlike `main`, there can be more than one `init` — in a single file they run in declaration order; across multiple files, in lexicographic filename order.

`init` is optional and typically used for global setup — establishing a database connection, loading configuration, setting environment variables, and so on.

### Defer

Postpones a function call until the surrounding function returns:

```go
func main() {
    defer fmt.Println("I am finished")
    fmt.Println("Doing some work...")
}
```

Multiple `defer`s form a **stack** — last in, first out:

```go
func main() {
    defer fmt.Println("I am finished")
    defer fmt.Println("Are you?")

    fmt.Println("Doing some work...")
}
```

```
Doing some work...
Are you?
I am finished
```

`defer` is commonly used for cleanup and error handling — closing a file, releasing a lock, and so on. (Functions can also use generics — covered in [Generics](#generics).)

---

## Modules

A **module** is a collection of Go packages stored in a file tree with a `go.mod` file at its root.

Go modules were introduced in Go 1.11 and have been the default (not experimental) since Go 1.13.

> **📌 Note on `GOPATH`:** older Go tutorials reference `GOPATH` — a variable defining your workspace root, with `src/` (source code), `pkg/` (compiled package code), and `bin/` (compiled binaries) subfolders. Modules mode doesn't require your code to live under `GOPATH/src` anymore — this is mostly historical context at this point, not something you need to actively manage.

```bash
go mod init example
```

If you plan to publish the module, the module path typically corresponds to its repository:

```bash
go mod init github.com/you/example
```

`go.mod` defines the module path, the Go version, and its dependency requirements:

```
module <name>

go <version>

require (
    ...
)
```

Add a dependency:

```bash
go get github.com/rs/zerolog
```

> **📌 Updated:** the original course used `go install` to add a dependency to `go.mod`. Use **`go get`** for that today — `go get` manages `go.mod`/`go.sum` entries (adding, upgrading, downgrading, removing versions), while `go install` builds and installs a binary into `$GOPATH/bin` and, since Go 1.18, no longer touches `go.mod` at all. They now do genuinely different jobs, not two names for the same thing.

`go.sum` (created alongside `go.mod`) records the expected cryptographic hashes of your dependencies' content, so builds are reproducible and tamper-evident.

List dependencies:

```bash
go list -m all
```

Remove unused ones:

```bash
go mod tidy
```

### Vendoring

Vendoring copies your third-party dependencies into your own repository:

```bash
go mod vendor
```

```
├── go.mod
├── go.sum
├── main.go
└── vendor
    ├── github.com
    │   └── rs
    │       └── zerolog
    │           └── ...
    └── modules.txt
```

---

## Packages

A **package** is a directory containing one or more Go source files. Every Go source file belongs to a package, declared at the top:

```go
package <package_name>
```

By convention, an executable program (one with `package main`, plus a `main()` entry point) is called a **command**; anything else is simply called a **package**.

### Imports and Exports

Any value (a variable, function, type) is **exported** — visible from other packages — if its identifier starts with an uppercase letter:

```go
package custom

var value int = 10 // not exported
var Value int = 20 // exported
```

Importing your own package, using the module path from `go.mod`:

```
---go.mod---
module example

go 1.24

---main.go---
package main

import "example/custom"

func main() {
    custom.Value
}
```

The package name is the last segment of the import path. Multiple imports:

```go
import (
    "fmt"

    "example/custom"
)
```

Aliased imports, to avoid collisions:

```go
import (
    "fmt"

    abcd "example/custom"
)
```

### External Dependencies

```bash
go get github.com/rs/zerolog
```

```go
package main

import (
    "github.com/rs/zerolog/log"

    abcd "example/custom"
)

func main() {
    log.Print(abcd.Value)
}
```

Check a package's own documentation (usually its README, or via `go doc`) before depending on it.

Go doesn't enforce a particular folder-structure convention — organize packages in whatever way is simple and intuitive for your project. (This project's own conventions are in [`design-system.md`](design-system.md).)

---

## Workspaces

Multi-module workspaces (Go 1.18+) let you work across multiple modules simultaneously without editing each module's `go.mod`. Every module in a workspace is treated as a root module for dependency resolution.

```bash
mkdir workspaces && cd workspaces
mkdir hello && cd hello
go mod init hello
```

```go
package main

import (
    "fmt"

    "golang.org/x/example/stringutil"
)

func main() {
    result := stringutil.Reverse("Hello Workspace")
    fmt.Println(result)
}
```

```bash
go get golang.org/x/example
go run main.go
```

```
ecapskroW olleH
```

Before workspaces, modifying a dependency's source required a `replace` directive in `go.mod`. With a workspace:

```bash
cd ..
go work init
go work use ./hello
```

```
go 1.24

use ./hello
```

Clone and modify the dependency locally:

```bash
git clone https://go.googlesource.com/example
```

```go
// example/stringutil/reverse.go
func Reverse(s string) string {
    return fmt.Sprintf("I can do whatever!! %s", s)
}
```

Add it to the workspace:

```bash
go work use ./example
```

```
go 1.24

use (
    ./example
    ./hello
)
```

```bash
go run hello
```

```
I can do whatever!! Hello Workspace
```

A niche but genuinely useful feature when you're actively developing a dependency alongside the code that uses it.

---

## Useful Commands

`go fmt` — formats source code. Formatting is enforced by convention across the whole ecosystem, so you spend zero time on style debates:

```bash
go fmt ./...
```

`go vet` — reports likely mistakes in your code:

```bash
go vet ./...
```

`go env` — prints Go's environment/build configuration.

`go doc` — shows documentation for a package or symbol:

```bash
go doc -src fmt Printf
```

`go help` lists everything else, including:

- `go fix` — rewrites programs using old APIs to use newer ones.
- `go generate` — runs code generation directives (see [`coding-standards.md`](coding-standards.md) for how this project uses it).
- `go install` — builds and installs a binary.
- `go clean` — removes files generated by the build.
- `go build` and `go test` — covered next, and in [Testing](#testing).

---

## Build

```go
package main

import "fmt"

func main() {
    fmt.Println("I am a binary!")
}
```

```bash
go build
```

Produces a binary named after the module. Specify the output explicitly:

```bash
go build -o app
./app
```

```
I am a binary!
```

### GOOS and GOARCH

Cross-compile for a different OS/architecture:

```bash
go tool dist list   # every supported target
```

```bash
GOOS=windows GOARCH=amd64 go build -o app.exe
```

### CGO_ENABLED

Controls CGO (calling C code from Go). Disabling it produces a fully statically linked binary with no external runtime dependencies — the standard choice for a minimal Docker image (this project's own services build this way; see [`architecture.md`](architecture.md)):

```bash
CGO_ENABLED=0 go build -o app
```

---

## Pointers

A pointer is a variable that stores the memory address of another variable.

```go
var x *T // T is the pointed-to type: int, string, etc.
```

```go
func main() {
    var p *int
    fmt.Println(p)
}
```

```
<nil>
```

`nil` is Go's predeclared identifier representing the zero value for pointers, interfaces, channels, maps, and slices.

Assign a value via the `&` (address-of) operator:

```go
func main() {
    a := 10
    var p *int = &a

    fmt.Println("address:", p)
}
```

```
0xc0000b8000
```

### Dereferencing

The `*` (asterisk) operator retrieves — or sets — the value a pointer points to:

```go
func main() {
    a := 10
    var p *int = &a

    fmt.Println("address:", p)
    fmt.Println("value:", *p)

    *p = 20
    fmt.Println("after:", a) // 20 — mutated through the pointer
}
```

### Pointers as Function Arguments

Pass by reference when a function needs to mutate the caller's data:

```go
myFunction(&a)

func myFunction(ptr *int) {}
```

### The `new` Function

Allocates memory for a value of the given type and returns a pointer to it:

```go
func main() {
    p := new(int)
    *p = 100

    fmt.Println("value", *p)
    fmt.Println("address", p)
}
```

### Pointer to a Pointer

```go
func main() {
    p := new(int)
    *p = 100

    p1 := &p

    fmt.Println("P value", *p, "address", p)
    fmt.Println("P1 value", *p1, "address", p)
    fmt.Println("Dereferenced value", **p1)
}
```

Go pointers **don't** support pointer arithmetic like C/C++:

```go
p1 := p * 2 // compiler error: invalid operation
```

But two pointers of the same type can be compared for equality:

```go
p := &a
p1 := &a

fmt.Println(p == p1)
```

### Why Pointers?

There's no single answer — pointers let you mutate data efficiently without copying large amounts of it. If you're coming from a language with no notion of pointers, give yourself time to build a mental model; it clicks with practice.

---

## Structs

A struct is a user-defined type grouping named fields into a single unit. If you're coming from an OO background, think of structs as lightweight, composition-friendly types without inheritance.

### Defining

```go
type Person struct {
    FirstName string
    LastName  string
    Age       int
}
```

Fields sharing a type can be collapsed:

```go
type Person struct {
    FirstName, LastName string
    Age                 int
}
```

### Declaring and Initializing

```go
func main() {
    var p1 Person
    fmt.Println("Person 1:", p1)
}
```

```
Person 1: { 0}
```

Every field is initialized to its zero value.

**Struct literal:**

```go
var p2 = Person{FirstName: "Karan", LastName: "Pratap Singh", Age: 22}
```

For readability, split across lines (trailing comma required):

```go
var p2 = Person{
    FirstName: "Karan",
    LastName:  "Pratap Singh",
    Age:       22,
}
```

Partial initialization — omitted fields get their zero value:

```go
var p3 = Person{
    FirstName: "Tony",
    LastName:  "Stark",
}
// Age defaults to 0
```

**Without field names** — positional, but *every* field must be provided:

```go
var p4 = Person{"Bruce", "Wayne", 40}
```

**Anonymous structs:**

```go
var a = struct {
    Name string
}{"Golang"}
```

### Accessing Fields

```go
p := Person{FirstName: "Karan", LastName: "Pratap Singh", Age: 22}
fmt.Println("FirstName", p.FirstName)
```

Via a pointer — no explicit dereference needed, `(*ptr).FirstName` and `ptr.FirstName` are equivalent:

```go
ptr := &p
fmt.Println((*ptr).FirstName)
fmt.Println(ptr.FirstName)
```

Or via `new`:

```go
p := new(Person)
p.FirstName = "Karan"
```

Two structs are equal if all corresponding fields are equal:

```go
p1 := Person{"a", "b", 20}
p2 := Person{"a", "b", 20}
fmt.Println(p1 == p2) // true
```

### Exported Fields

Same capitalization rule as everything else — a lowercase field/struct name is unexported, visible only within its own package:

```go
type Person struct {
    FirstName, LastName string
    Age                 int
    zipCode             string // not exported
}
```

### Embedding and Composition

Go doesn't support inheritance, but embedding gets you something similar:

```go
type Person struct {
    FirstName, LastName string
    Age                 int
}

type SuperHero struct {
    Person
    Alias string
}
```

```go
s := SuperHero{}
s.FirstName = "Bruce"
s.Alias = "batman"
```

Embedding is usually not recommended for anything beyond the simplest cases — **composition is generally preferred**. Define the field explicitly instead:

```go
type SuperHero struct {
    Person Person
    Alias  string
}
```

```go
p := Person{"Bruce", "Wayne", 40}
s := SuperHero{p, "batman"}
```

There's no universally right answer here, but composition tends to age better as a codebase grows.

### Struct Tags

Attach metadata to a field, readable via the `reflect` package — most commonly used by encoding packages (JSON, XML, YAML), ORMs, and configuration libraries:

```go
type Animal struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
```

(See [`coding-standards.md §11`](coding-standards.md) for how this project uses JSON struct tags.)

### Properties

**Structs are value types.** Assigning or passing a struct copies it:

```go
type Point struct{ X, Y float64 }

p1 := Point{1, 2}
p2 := p1 // copy
p2.X = 2

fmt.Println(p1) // {1 2}
fmt.Println(p2) // {2 2}
```

An empty struct occupies **zero bytes** of storage — useful as a signal-only value (e.g. in a set implemented as `map[T]struct{}`, or a channel used purely for signaling):

```go
var s struct{}
fmt.Println(unsafe.Sizeof(s)) // 0
```

---

## Methods

Go is not object-oriented — no classes, no inheritance — but it does have **methods**: a function with a special *receiver* argument.

```go
func (variable T) Name(params) (returnTypes) {}
```

```go
type Car struct {
    Name string
    Year int
}

func (c Car) IsLatest() bool {
    return c.Year >= 2017
}
```

```go
c := Car{"Tesla", 2021}
fmt.Println("IsLatest", c.IsLatest())
```

### Pointer Receivers

A **value receiver** operates on a copy — mutations inside the method aren't visible to the caller:

```go
func (c Car) UpdateName(name string) {
    c.Name = name // no effect on the caller's Car
}
```

Switch to a **pointer receiver** to actually mutate the caller's value:

```go
func (c *Car) UpdateName(name string) {
    c.Name = name
}
```

```go
c := Car{"Tesla", 2021}
c.UpdateName("Toyota")
fmt.Println("Car:", c) // {Toyota 2021}
```

### Properties

Pointer-receiver method calls are syntactic sugar — `c.UpdateName(...)` is really `(&c).UpdateName(...)`.

The receiver's variable name can be omitted if unused:

```go
func (Car) UpdateName(...) {}
```

Methods work on non-struct types too:

```go
type MyInt int

func (i MyInt) isGreater(value int) bool {
    return i > MyInt(value)
}
```

### Why Methods Instead of Functions?

There's no absolute answer — but methods let you reuse the same method name across multiple receiver types without a naming collision, since a method is tied to its type. Beyond that, it often comes down to what reads more clearly for the code at hand.

---

## Arrays and Slices

### Arrays

An array is a **fixed-size** collection of same-typed elements, stored sequentially:

```go
var a [n]T
```

```go
func main() {
    var arr [4]int
    fmt.Println(arr) // [0 0 0 0] — zero-valued
}
```

**Initialization:**

```go
arr := [4]int{1, 2, 3, 4}
```

**Access:**

```go
fmt.Println(arr[0]) // 1
```

**Iteration** — via index and `len`, or via `range`:

```go
for i := 0; i < len(arr); i++ {
    fmt.Printf("Index: %d, Element: %d\n", i, arr[i])
}
```

```go
for i, e := range arr {
    fmt.Printf("Index: %d, Element: %d\n", i, e)
}
```

`range` is versatile:

```go
for i, e := range arr {} // index + element
for _, e := range arr {} // element only
for i := range arr {}    // index only
for range arr {}         // just loop
```

**Multi-dimensional arrays:**

```go
arr := [2][4]int{
    {1, 2, 3, 4},
    {5, 6, 7, 8},
}
```

Let the compiler infer the length with `...`:

```go
arr := [...][4]int{
    {1, 2, 3, 4},
    {5, 6, 7, 8},
}
```

### Properties

An array's length is part of its **type** — `[4]int` and `[2]int` are distinct, incompatible types, and arrays can't be resized (resizing would mean changing the type):

```go
var a = [4]int{1, 2, 3, 4}
var b [2]int = a // ❌ cannot use a (type [4]int) as type [2]int
```

Arrays are **value types**, unlike C/C++/Java — assigning or passing one copies the whole thing:

```go
a := [7]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
b := a // copy

b[0] = "Monday"

fmt.Println(a) // [Mon Tue Wed Thu Fri Sat Sun] — unaffected
fmt.Println(b) // [Monday Tue Wed Thu Fri Sat Sun]
```

### Slices

A slice is a segment of an array — more powerful and flexible than an array's fixed size.

A slice is really three things: a pointer to an underlying array, a length, and a capacity (the maximum size the segment can grow to before reallocating).

```go
a := [5]int{20, 15, 5, 30, 25}
s := a[1:4]

fmt.Printf("Array: %v, Length: %d, Capacity: %d\n", a, len(a), cap(a))
// Array: [20 15 5 30 25], Length: 5, Capacity: 5

fmt.Printf("Slice: %v, Length: %d, Capacity: %d", s, len(s), cap(s))
// Slice: [15 5 30], Length: 3, Capacity: 4
```

**Declaration** — no length required:

```go
var s []string
fmt.Println(s == nil) // true — a slice's zero value is nil, unlike an array
```

**Initialization**, via `make`:

```go
make([]T, len, cap) []T
```

```go
s := make([]string, 0, 0)
```

Via a slice literal:

```go
s := []string{"Go", "TypeScript"}
```

Or carved from an existing array (or another slice) with `a[low:high]`:

```go
a := [4]string{"C++", "Go", "Java", "TypeScript"}

s1 := a[0:2] // [C++ Go]
s2 := a[:3]  // [C++ Go Java] — missing low implies 0
s3 := a[2:]  // [Java TypeScript] — missing high implies len(a)
```

**Iteration** — same as arrays: `len` + index, or `range`.

### Slice Functions

**`copy`** — copies elements from a source slice into a destination slice, returning the count copied:

```go
func copy(dst, src []T) int
```

```go
s1 := []string{"a", "b", "c", "d"}
s2 := make([]string, len(s1))

n := copy(s2, s1)
fmt.Println("Elements:", n) // 4
```

**`append`** — appends elements, returning a new slice:

```go
func append(slice []T, elems ...T) []T
```

```go
s1 := []string{"a", "b", "c", "d"}
s2 := append(s1, "e", "f")

fmt.Println("s1:", s1) // [a b c d] — unchanged
fmt.Println("s2:", s2) // [a b c d e f]
```

If the slice doesn't have enough capacity, `append` allocates a new, bigger underlying array and copies everything over before appending.

### Properties

**Slices are reference types**, unlike arrays — mutating a slice's elements mutates the underlying array too:

```go
a := [7]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
s := a[0:2]

s[0] = "Sun"

fmt.Println(a) // [Sun Tue Wed Thu Fri Sat Sun]
fmt.Println(s) // [Sun Tue]
```

Slices work with variadic parameters via `...`:

```go
values := []int{1, 2, 3}
sum := add(values...)
```

---

## Maps

A **map** is an unordered collection of key-value pairs — keys are unique, values may not be. Used for fast lookup, insertion, and deletion by key.

### Declaration

```go
var m map[K]V
```

```go
var m map[string]int
fmt.Println(m) // nil — a map's zero value is nil
```

A `nil` map has no keys, and attempting to *write* to one causes a runtime panic (reading from a `nil` map is safe and returns zero values).

### Initialization

Via `make`:

```go
m := make(map[string]int)
```

Via a map literal (trailing comma required):

```go
m := map[string]int{
    "a": 0,
    "b": 1,
}
```

With custom types — the value type can be omitted when it's inferable:

```go
type User struct{ Name string }

m := map[string]User{
    "a": {"Peter"},
    "b": {"Seth"},
}
```

### Add, Retrieve, Exists, Update, Delete

```go
m["c"] = User{"Steve"}          // add
c := m["c"]                     // retrieve — {Steve}
d := m["d"]                     // missing key → zero value, {}

c, ok := m["c"]                 // ok = true
d, ok := m["d"]                 // ok = false

m["a"] = User{"Roger"}          // update

delete(m, "a")                  // delete — no-op if the key doesn't exist
```

### Iteration

```go
for key, value := range m {
    fmt.Printf("Key: %s, Value: %v\n", key, value)
}
```

A map is unordered — iteration order is **not** guaranteed to be stable across runs.

### Properties

**Maps are reference types.** Assigning a map to a new variable makes both variables refer to the same underlying data:

```go
m1 := map[string]User{"a": {"Peter"}}
m2 := m1
m2["c"] = User{"Steve"}

fmt.Println(m1) // map[a:{Peter} c:{Steve}] — m1 sees the change too
```

---

## Interfaces

An interface is an abstract type defined by a set of method signatures — it describes *behavior*, not data.

The classic example: a power socket. Different devices (`mobile`, `laptop`, `toaster`, `kettle`) all need to plug into the same `socket`.

Without an interface, `socket.Plug` would need a separate method per device type — unworkable as device types grow. Instead, define the shared behavior:

```go
type PowerDrawer interface {
    Draw(power int)
}
```

By convention, single-method interfaces are named with an `-er` suffix.

```go
func (socket) Plug(device PowerDrawer, power int) {
    device.Draw(power)
}
```

Any type that implements `Draw(power int)` satisfies the interface:

```go
type mobile struct{ brand string }

func (m mobile) Draw(power int) {
    fmt.Printf("%T -> brand: %s, power: %d\n", m, m.brand, power)
}

type laptop struct{ cpu string }

func (l laptop) Draw(power int) {
    fmt.Printf("%T -> cpu: %s, power: %d\n", l, l.cpu, power)
}
```

```go
s := socket{}
s.Plug(mobile{"Apple"}, 10)
s.Plug(laptop{"Intel i9"}, 50)
```

`socket` never needs to know anything about a specific device type — adding a new device type means adding a `Draw` method to it, not touching `socket` at all. This is the actual power of the interface: it decouples the two sides.

**Go interfaces are satisfied implicitly** — no `implements` keyword. A type satisfies an interface automatically the moment it has every method the interface requires.

### The Empty Interface

`interface{}` (or, since Go 1.18, the built-in alias `any`) can hold a value of *any* type:

```go
var x any
```

Useful for handling values of unknown type — reading heterogeneous data from an API, or `fmt.Println`'s own variadic parameter.

### Type Assertion

Access an interface value's underlying concrete value:

```go
var i any = "hello"
s := i.(string)
```

The two-value form tests whether the assertion holds, without panicking on failure:

```go
s, ok := i.(string) // "hello", true
f, ok := i.(float64) // 0, false — zero value, no panic
```

The single-value form **panics** if the interface doesn't hold the asserted type:

```go
f := i.(float64) // panic: interface conversion: interface {} is string, not float64
```

### Type Switch

```go
var t any = "hello"

switch t := t.(type) {
case string:
    fmt.Printf("string: %s\n", t)
case bool:
    fmt.Printf("boolean: %v\n", t)
case int:
    fmt.Printf("integer: %d\n", t)
default:
    fmt.Printf("unexpected: %T\n", t)
}
```

### Properties

The zero value of an interface is `nil`:

```go
var i MyInterface
fmt.Println(i) // <nil>
```

Interfaces can be embedded, same as structs:

```go
type interface3 interface {
    interface1
    interface2
}
```

Interface values are comparable, and can be thought of as a `(value, concrete type)` tuple under the hood:

```go
var i MyInterface = MyType{10}
fmt.Printf("(%v, %T)\n", i, i) // ({10}, main.MyType)
```

> "The bigger the interface, the weaker the abstraction." — Rob Pike. See this project's own [Interface Design conventions](coding-standards.md#5-interfaces--constructors) for how that principle applies here specifically: small, consumer-declared interfaces, not broad producer-defined ones.

---

## Errors

Go has **errors**, not exceptions — there's no exception handling in the language. An error is just a value of the built-in `error` interface type:

```go
type error interface {
    Error() string
}
```

```go
func Divide(a, b int) int {
    return a / b // panics on b == 0 — not what we want
}
```

### Constructing Errors

**`errors.New`:**

```go
import "errors"

func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("cannot divide by zero")
    }
    return a / b, nil
}
```

A `nil` error means no error occurred — `error` is an interface, and `nil` is its zero value.

```go
result, err := Divide(4, 0)

if err != nil {
    fmt.Println(err)
    return
}

fmt.Println(result)
```

**`fmt.Errorf`** — like `fmt.Sprintf`, but returns an `error`. Commonly used to add context to an error as it propagates up:

```go
func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, fmt.Errorf("cannot divide %d by zero", a)
    }
    return a / b, nil
}
```

> See [`coding-standards.md §3`](coding-standards.md) for this project's actual error-wrapping conventions (`%w` vs `%v`, when to wrap at all) — the reasoning there is more specific than what's covered in this general course.

### Sentinel Errors

Predeclared errors that calling code can check for explicitly:

```go
var ErrDivideByZero = errors.New("cannot divide by zero")

func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, ErrDivideByZero
    }
    return a / b, nil
}
```

Convention: prefix with `Err` (e.g. `ErrNotFound`). Check for a specific sentinel with `errors.Is`:

```go
result, err := Divide(4, 0)

if err != nil {
    switch {
    case errors.Is(err, ErrDivideByZero):
        fmt.Println(err)
    default:
        fmt.Println("no idea!")
    }
    return
}
```

### Custom Errors

Since `error` is just an interface, any type with an `Error() string` method qualifies:

```go
type DivisionError struct {
    Code int
    Msg  string
}

func (d DivisionError) Error() string {
    return fmt.Sprintf("code %d: %s", d.Code, d.Msg)
}

func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, DivisionError{Code: 2000, Msg: "cannot divide by zero"}
    }
    return a / b, nil
}
```

Use `errors.As` (not `errors.Is`) to check whether an error is a specific *type* and unwrap it into a typed variable:

```go
result, err := Divide(4, 0)

if err != nil {
    var divErr DivisionError

    switch {
    case errors.As(err, &divErr):
        fmt.Println(divErr)
    default:
        fmt.Println("no idea!")
    }
    return
}
```

**`errors.Is` checks whether an error *is* a specific error value; `errors.As` checks whether it *can be converted to* a specific error type.** A bare type assertion (`err.(DivisionError)`) also works but isn't preferred — it doesn't unwrap through a chain of wrapped errors the way `errors.As` does.

Go's explicit `if err != nil` idiom is a deliberate design choice, not an oversight — it forces you to actually handle an error at the point it occurs, which tends to produce more readable, more honest code than an implicit try/catch.

---

## Panic and Recover

Errors handle most abnormal conditions, but some situations mean the program genuinely cannot continue — that's what `panic` is for.

### Panic

```go
func panic(any)
```

Stops the current goroutine's normal execution immediately; control unwinds up the call stack until the program exits with the panic message and a stack trace (unless something `recover`s it first — see below).

```go
func WillPanic() {
    panic("Woah")
}
```

```
panic: Woah

goroutine 1 [running]:
main.WillPanic(...)
        .../main.go:8
main.main()
        .../main.go:4 +0x38
exit status 2
```

### Recover

`recover`, combined with `defer`, regains control of a panicking goroutine:

```go
func recover() any
```

```go
func handlePanic() {
    data := recover()
    fmt.Println("Recovered:", data)
}

func WillPanic() {
    defer handlePanic()
    panic("Woah")
}
```

```
Recovered: Woah
```

`panic`/`recover` resemble try/catch, but the convention in Go is to **avoid** them and use errors wherever possible.

### When Panic Is Actually Appropriate

1. **An unrecoverable error** — e.g. a required configuration file fails to load at startup; there's nothing meaningful left to do.
2. **A programmer error** — e.g. dereferencing a `nil` pointer. This should never happen in correct code; a panic surfaces the bug loudly instead of silently continuing with bad state.

(See [`coding-standards.md §3`](coding-standards.md) for exactly where this project draws that line in practice.)

---

## Testing

```go
// math/add.go
package math

func Add(a, b int) int {
    return a + b
}
```

```go
package main

import (
    "example/math"
    "fmt"
)

func main() {
    fmt.Println(math.Add(2, 2)) // 4
}
```

Test files use a `_test.go` suffix:

```
.
├── go.mod
├── main.go
└── math
    ├── add.go
    └── add_test.go
```

Using a separate `math_test` package (rather than `math` itself) keeps tests decoupled from implementation:

```go
package math_test

import "testing"

func TestAdd(t *testing.T) {}
```

Run tests with `go test`, not `go run`:

```bash
go test ./math      # by package
go test ./...        # everything; reports "no test files" where there are none
```

A real test — compare the result to an expectation, fail via `t.Fail()`/`t.Errorf()`:

```go
package math_test

import (
    "example/math"
    "testing"
)

func TestAdd(t *testing.T) {
    got := math.Add(1, 1)
    expected := 2

    if got != expected {
        t.Errorf("Expected %d but got %d", expected, got)
    }
}
```

Test results are cached — if you change test *inputs* without changing code, you may need `go clean -testcache` to force a re-run.

### Table-Driven Tests

Define cases as data, iterate over them — the standard Go idiom for testing multiple inputs against one function (also this project's own default, per [`coding-standards.md §13`](coding-standards.md)):

```go
package math_test

import (
    "example/math"
    "testing"
)

type addTestCase struct {
    a, b, expected int
}

var testCases = []addTestCase{
    {1, 1, 2},
    {25, 25, 50},
    {2, 1, 3},
    {1, 10, 11},
}

func TestAdd(t *testing.T) {
    for _, tc := range testCases {
        got := math.Add(tc.a, tc.b)
        if got != tc.expected {
            t.Errorf("Expected %d but got %d", tc.expected, got)
        }
    }
}
```

`addTestCase` is lowercase deliberately — it has no reason to be exported outside the test file.

> **📌 Updated — prefer `t.Run` per case.** The idiomatic modern form names each case and runs it as its own subtest, so a failure identifies exactly which case broke instead of a bare line number:
>
> ```go
> func TestAdd(t *testing.T) {
>     for _, tc := range testCases {
>         t.Run(tc.name, func(t *testing.T) {
>             got := math.Add(tc.a, tc.b)
>             if got != tc.expected {
>                 t.Errorf("Expected %d but got %d", tc.expected, got)
>             }
>         })
>     }
> }
> ```

### Code Coverage

```bash
go test ./math -coverprofile=coverage.out
go tool cover -html=coverage.out   # HTML report
```

### Fuzz Testing

Introduced in Go 1.18 — automatically manipulates inputs to find edge cases a human would likely miss:

```go
func FuzzTestAdd(f *testing.F) {
    f.Fuzz(func(t *testing.T, a, b int) {
        math.Add(a, b)
    })
}
```

```bash
go test -fuzz FuzzTestAdd example/math
```

If `Add` had a bug — say, panicking whenever `a > b + 10` — fuzzing would surface the exact failing input automatically, something a hand-written table of cases would likely never think to try.

---

## Generics

Introduced in Go 1.18 — **generics** mean parameterized types: code where the concrete type is decided by the caller, not hardcoded.

Without generics, you'd write one function per type even when the logic is identical:

```go
func sumInt(a, b int) int { return a + b }
func sumFloat(a, b float64) float64 { return a + b }
func sumString(a, b string) string { return a + b }
```

A generic function:

```go
func fnName[T constraint]() { ... }
```

`T` is the type parameter; `constraint` is the interface describing which types are allowed.

```go
func sum[T any](a, b T) T {
    fmt.Println(a, b)
    return a // placeholder — see below
}
```

`any` (Go 1.18+) is an alias for the empty interface `interface{}`.

Explicit type arguments work, but **type inference** usually makes them unnecessary:

```go
sum[int](1, 2)   // explicit
sum(1, 2)         // inferred — the common case
```

But `any` alone doesn't support operators:

```go
func sum[T any](a, b T) T {
    return a + b // ❌ invalid operation: operator + not defined on T
}
```

Define a constraint that limits `T` to types the operator actually supports:

```go
type SumConstraint interface {
    int | float64 | string
}

func sum[T SumConstraint](a, b T) T {
    return a + b
}
```

```go
fmt.Println(sum(1, 2))     // 3
fmt.Println(sum(4.0, 2.0)) // 6
fmt.Println(sum("a", "b")) // ab
```

`~int` (the tilde token) means "any type whose *underlying* type is `int`" — this is what lets a constraint match your own named types too, not just the built-in ones:

```go
type Signed interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64
}
```

> **📌 Updated — you rarely need to hand-write these constraints today.** The original course has you `go get golang.org/x/exp/constraints` for `Ordered` and friends. As of **Go 1.21**, the standard library ships `cmp.Ordered` directly — no external dependency required:
>
> ```go
> import "cmp"
>
> func sum[T cmp.Ordered](a, b T) T {
>     return a + b
> }
> ```
>
> Go 1.21 also added the standard-library **`slices`** and **`maps`** packages, providing generic implementations of the operations you'd otherwise hand-roll constantly — `slices.Contains`, `slices.Sort`, `slices.Index`, `maps.Keys`, `maps.Equal`, and more. Before reaching for a generic helper of your own, check whether `slices`/`maps`/`cmp` already has it.

### When to Use Generics

- Functions operating generically over arrays, slices, maps, or channels.
- General-purpose data structures — a stack, a linked list.
- Genuinely reducing duplication.

Use them sparingly. The Go team's own advice: start concrete, and only reach for a generic version once you've written near-identical code two or three times.

---

## Concurrency

Go's most distinctive feature. Start with the definition:

**Concurrency** is the ability to break a program into independently-executable parts. The final result is the same as running everything sequentially — concurrency is about *structure*, not necessarily about speed on its own.

### Concurrency vs. Parallelism

People often conflate the two, but they're different concepts. Concurrency is managing multiple things at once; parallelism is actually *running* multiple things at once.

> "Concurrency is about dealing with lots of things at once. Parallelism is about doing lots of things at once." — Rob Pike

Go's concurrency model is **CSP** (Communicating Sequential Processes), a formalism from Tony Hoare (1978) describing how independent processes interact by communicating over channels, rather than by sharing memory directly. Go and Erlang both draw heavily on it. You'll see this play out concretely in [Goroutines](#goroutines) and [Channels](#channels).

### Basic Concepts

- **Data race** — two goroutines access the same resource concurrently, at least one of them writing.
- **Race condition** — the timing or ordering of events changes the correctness of the result.
- **Deadlock** — every process is blocked waiting on another; nothing can proceed. Formally requires all four **Coffman conditions** to hold simultaneously:
  - **Mutual exclusion** — a resource is held exclusively by one process at a time.
  - **Hold and wait** — a process holds one resource while waiting on another.
  - **No preemption** — a resource can only be released voluntarily by whoever holds it.
  - **Circular wait** — a cycle of processes, each waiting on the next.
- **Livelock** — processes are actively doing work, but none of it moves the program's state forward.
- **Starvation** — a process is perpetually denied the resources it needs to proceed, often due to an unfair scheduling policy or a lurking deadlock.

---

## Goroutines

> "Don't communicate by sharing memory; share memory by communicating." — Rob Pike

A **goroutine** is a lightweight thread of execution managed by the Go runtime — not an OS thread. `main` itself runs as a goroutine.

The Go runtime scheduler multiplexes many goroutines onto a small number of OS threads using cooperative scheduling — when a goroutine blocks or finishes, the scheduler moves others onto the freed thread, so nothing sits blocked forever waiting for a thread.

Turn any function call into a goroutine with `go`:

```go
go fn(x, y, z)
```

This follows the **fork-join model**: a child "forks" off the parent to run concurrently, then "joins" back once it completes.

```go
func speak(arg string) {
    fmt.Println(arg)
}

func main() {
    go speak("Hello World")
}
```

Run this and you'll likely see **no output at all** — `main` (and the whole program) exits before the goroutine gets a chance to run. `time.Sleep` "fixes" it, but only by accident, not by design:

```go
func main() {
    go speak("Hello World")
    time.Sleep(1 * time.Second) // don't actually do this — see Channels/sync.WaitGroup instead
}
```

The real fix — actually waiting for a goroutine, not guessing how long it'll take — is a channel ([next section](#channels)) or a `sync.WaitGroup` (see [Sync Package](#sync-package)). Since goroutines share the same address space, any memory they access concurrently needs to be synchronized — that's the whole subject of the rest of this section.

---

## Channels

A channel is a communication pipe between goroutines — values go in one end and come out the other, in order, until the channel is closed. Channels are the concrete mechanism behind Go's CSP model.

### Creating a Channel

```go
var ch chan T
```

```go
var ch chan string
fmt.Println(ch) // <nil> — zero value; sending on a nil channel blocks forever
```

Initialize with `make`, same as slices/maps:

```go
ch := make(chan string)
```

### Sending and Receiving

```go
func speak(arg string, ch chan string) {
    ch <- arg // send
}

func main() {
    ch := make(chan string)

    go speak("Hello World", ch)

    data := <-ch // receive
    fmt.Println(data)
}
```

This is the correct way to solve the goroutine-timing problem from the previous section — the receive on `<-ch` genuinely blocks until the goroutine sends, no guessing required.

### Buffered Channels

Accept a limited number of values without a matching receiver already waiting — specify capacity as the second `make` argument:

```go
ch := make(chan string, 2)

go speak("Hello World", ch)
go speak("Hi again", ch)

data1 := <-ch
data2 := <-ch
```

A send to a buffered channel only blocks once the buffer is full; a receive only blocks once it's empty. An unbuffered channel (the default) has capacity 0.

### Directional Channels

Restrict a channel parameter to send-only or receive-only — increases type safety, since a plain `chan T` can do both by default:

```go
func speak(arg string, ch chan<- string) { // send-only
    ch <- arg
}
```

### Closing Channels

```go
close(ch)
```

Test whether a channel is closed via the two-value receive form:

```go
data, ok := <-ch
```

`ok` is `false` once the channel is closed and drained — analogous to the "does this key exist" check on a map.

### Properties

- A send to a `nil` channel blocks forever.
- A receive from a `nil` channel blocks forever.
- A send to a **closed** channel panics.
- A receive from a closed channel returns the zero value immediately (never blocks).

```go
c := make(chan int, 2)
c <- 5
c <- 4
close(c)

for i := 0; i < 4; i++ {
    fmt.Printf("%d ", <-c) // 5 4 0 0
}
```

**`range` over a channel** — receives until the channel is closed:

```go
ch := make(chan string, 2)
ch <- "Hello"
ch <- "World"
close(ch)

for data := range ch {
    fmt.Println(data)
}
```

---

## Select

`select` blocks and waits on multiple channel operations simultaneously — it runs the first case that's ready, choosing randomly if several are ready at once.

```go
one := make(chan string)
two := make(chan string)

go func() {
    time.Sleep(2 * time.Second)
    one <- "One"
}()

go func() {
    time.Sleep(1 * time.Second)
    two <- "Two"
}()

select {
case result := <-one:
    fmt.Println("Received:", result)
case result := <-two:
    fmt.Println("Received:", result)
}
```

A `default` case runs immediately if nothing else is ready — the way to send/receive without blocking:

```go
select {
case result := <-one:
    fmt.Println("Received:", result)
case result := <-two:
    fmt.Println("Received:", result)
default:
    fmt.Println("Default...")
}
```

An empty `select {}` blocks forever.

---

## Sync Package

Since goroutines share an address space, access to shared memory needs synchronization — `sync` provides the primitives for that.

### WaitGroup

Waits for a collection of goroutines to finish.

- `Add(n int)` — set how many goroutines to wait for (call *before* launching them).
- `Done()` — called inside a goroutine when it finishes (typically via `defer`).
- `Wait()` — blocks until every `Add`ed goroutine has called `Done`.

```go
var wg sync.WaitGroup

wg.Add(1)
go func() {
    defer wg.Done()
    work()
}()

wg.Wait()
```

Passed to a function directly, always by pointer:

```go
func work(wg *sync.WaitGroup) {
    defer wg.Done()
    fmt.Println("working...")
}

wg.Add(1)
go work(&wg)
wg.Wait()
```

**A `WaitGroup` must never be copied after first use** — pass it by pointer, always, or its internal counter gets corrupted.

### Mutex

A mutual-exclusion lock, preventing concurrent access to a **critical section** — code that touches shared state and must not run on more than one goroutine at a time.

- `Lock()` — acquire the lock.
- `Unlock()` — release it.
- `TryLock()` — attempt to lock without blocking, reporting success.

Without a mutex, concurrent updates race:

```go
type Counter struct{ value int }

func (c *Counter) Update(n int, wg *sync.WaitGroup) {
    defer wg.Done()
    c.value += n // data race — multiple goroutines touching this concurrently
}
```

With one:

```go
type Counter struct {
    m     sync.Mutex
    value int
}

func (c *Counter) Update(n int, wg *sync.WaitGroup) {
    defer wg.Done()
    c.m.Lock()
    c.value += n
    c.m.Unlock()
}
```

> Same rule as `WaitGroup`: a `Mutex` must never be copied after first use.

### RWMutex

A reader/writer lock — any number of readers can hold it simultaneously, but a writer needs exclusive access. Readers don't block each other, only writers block everyone. Prefer `RWMutex` over a plain `Mutex` when reads vastly outnumber writes.

- `Lock()` / `Unlock()` — exclusive (writer) lock.
- `RLock()` / `RUnlock()` — shared (reader) lock.

```go
type Counter struct {
    m     sync.RWMutex
    value int
}

func (c *Counter) Update(n int, wg *sync.WaitGroup) {
    defer wg.Done()
    c.m.Lock()
    c.value += n
    c.m.Unlock()
}

func (c *Counter) GetValue(wg *sync.WaitGroup) {
    defer wg.Done()
    c.m.RLock()
    defer c.m.RUnlock()
    fmt.Println("Get value:", c.value)
}
```

Both `Mutex` and `RWMutex` satisfy the `sync.Locker` interface:

```go
type Locker interface {
    Lock()
    Unlock()
}
```

### Cond

A condition variable, used to coordinate goroutines waiting on a shared resource's state change. Each `Cond` has an associated `Locker` (typically a `*Mutex`), held while checking/changing the condition and while calling `Wait`.

- `NewCond(l Locker)` — construct one.
- `Broadcast()` — wake every goroutine waiting on the condition.
- `Signal()` — wake exactly one, if any are waiting.
- `Wait()` — atomically unlock and block, then re-lock on wakeup.

```go
var done = false

func read(name string, c *sync.Cond) {
    c.L.Lock()
    for !done {
        c.Wait()
    }
    fmt.Println(name, "starts reading")
    c.L.Unlock()
}

func write(name string, c *sync.Cond) {
    time.Sleep(time.Second)

    c.L.Lock()
    done = true
    c.L.Unlock()

    c.Broadcast()
}
```

Real-world use case: several readers need to wait until a writer has finished producing data, and all of them need to be notified together — a channel alone can only wake one receiver, not broadcast to many.

### Once

Guarantees a function runs exactly once, no matter how many goroutines call it.

- `Do(f func())` — the only method. Only the first call actually invokes `f`.

```go
var once sync.Once

for i := 0; i < 100; i++ {
    go func() {
        once.Do(increment)
    }()
}
// count only ever increments once, regardless of 100 concurrent calls
```

### Pool

A scalable, concurrency-safe pool of temporary objects. Any pooled value can be dropped at any time without notice — the pool grows under load and shrinks when idle.

- `Get()` — remove and return an arbitrary item (or construct one via `New`, if configured).
- `Put(x any)` — return an item to the pool.

```go
var pool = sync.Pool{
    New: func() any {
        return &Person{}
    },
}

person := pool.Get().(*Person)
pool.Put(person)
```

`Pool` exists to relieve pressure on the garbage collector for high-churn temporary allocations — it's slower than a simple initialization, so use it where allocation pressure is a real, measured problem, not by default. Like `WaitGroup`/`Mutex`, never copy one after first use.

### Map

A drop-in-shaped alternative to `map[any]any` that's safe for concurrent use without an external lock.

Optimized specifically for two patterns: (1) a key is written once, read many times (an ever-growing cache), or (2) many goroutines read/write **disjoint** key sets. Outside those patterns, **a plain map plus your own `Mutex`/`RWMutex` is usually the better, more type-safe choice** — `sync.Map` is the exception, not the default.

- `Store(key, value any)`
- `Load(key any) (value any, ok bool)`
- `LoadOrStore`, `LoadAndDelete`, `Delete`, `Range(f func(key, value any) bool)`

```go
var m sync.Map

m.Store(1, "value 1")
v, ok := m.Load(1)
```

### Atomic

`sync/atomic` provides low-level atomic primitives for integers and pointers — the building blocks for synchronization algorithms, supporting add, load, store, swap, and compare-and-swap.

```go
func add(w *sync.WaitGroup, num *int32) {
    defer w.Done()
    atomic.AddInt32(num, 1)
}

var n int32
var wg sync.WaitGroup

wg.Add(1000)
for i := 0; i < 1000; i++ {
    go add(&wg, &n)
}
wg.Wait()

fmt.Println("Result:", n) // always exactly 1000 — the increment can't be interrupted
```

> **📌 Updated — Go 1.19 added typed atomics.** `atomic.Int32`, `atomic.Int64`, `atomic.Bool`, `atomic.Pointer[T]`, etc. wrap the old function-based API in a struct with methods (`.Add()`, `.Load()`, `.Store()`, `.CompareAndSwap()`), which reads more clearly and is harder to misuse (no risk of passing the wrong pointer type). Prefer the typed form in new code:
>
> ```go
> var n atomic.Int32
> n.Add(1)
> fmt.Println(n.Load())
> ```

---

## Advanced Concurrency Patterns

These patterns typically combine in real systems, not use one in isolation.

### Generator

Produces a sequence of values on demand via a channel — comparable to `yield` in Python/JavaScript:

```go
func generator() <-chan int {
    ch := make(chan int)

    go func() {
        for i := 0; ; i++ {
            ch <- i
        }
    }()

    return ch
}
```

```go
ch := generator()
for i := 0; i < 5; i++ {
    fmt.Println("Value:", <-ch)
}
```

### Fan-in

Combines multiple input channels into one output channel (multiplexing). Input ordering isn't guaranteed.

```go
func fanIn(inputs ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    out := make(chan int)

    wg.Add(len(inputs))
    for _, in := range inputs {
        go func(ch <-chan int) {
            defer wg.Done()
            for v := range ch {
                out <- v
            }
        }(in)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

### Fan-out

The inverse — splits one input channel across multiple output channels, distributing work to multiple uniform consumers:

```go
func fanOut(in <-chan int) <-chan int {
    out := make(chan int)

    go func() {
        defer close(out)
        for v := range in {
            out <- v
        }
    }()

    return out
}
```

> Fan-out is a different pattern from pub/sub — each value goes to exactly one output, not to all of them.

### Pipeline

A series of stages connected by channels, each stage a group of goroutines running the same function — receive from upstream, transform, send downstream. The first stage is the *source*/*producer*; the last is the *sink*/*consumer*.

```go
out := filter(in) // keep even numbers
out = square(out)  // square each
out = half(out)    // halve each
```

Each stage is independently swappable and composable — that separation of concerns is the actual point of the pattern, not just the pipe-and-filter shape itself.

### Worker Pool

Distributes work across a fixed number of concurrent workers — the pattern you reach for most often in practice:

```go
const totalWorkers = 2

jobs := make(chan int, totalJobs)
results := make(chan int, totalJobs)

for w := 1; w <= totalWorkers; w++ {
    go worker(w, jobs, results)
}

for j := 1; j <= totalJobs; j++ {
    jobs <- j
}
close(jobs)
```

`totalWorkers` is often set to `runtime.NumCPU()` — the number of logical CPUs available to the current process — as a sensible default, tuned from there based on whether the work is CPU-bound or I/O-bound.

### Queuing

Bounds how many items are processed at once, using a buffered channel purely as a semaphore (an empty `struct{}` costs zero bytes, so it's ideal for this — it carries no data, just a slot):

```go
const limit = 2

queue := make(chan struct{}, limit)

func process(work int, queue chan struct{}, wg *sync.WaitGroup) {
    queue <- struct{}{} // acquire a slot
    go func() {
        defer wg.Done()
        // do work
        <-queue // release the slot
    }()
}
```

### Other Patterns Worth Knowing

Tee channel, bridge channel, ring-buffer channel, bounded parallelism — not covered in depth here, but worth looking up once the patterns above feel comfortable.

---

## Context

In concurrent programs, you often need to preempt work because of a timeout, an explicit cancellation, or a failure elsewhere in the system. `context` carries request-scoped values, cancellation signals, and deadlines across API boundaries to every goroutine handling a request.

### Core Types

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}
```

- `Done()` — a channel closed on cancellation or timeout (`nil` if the context can never be canceled).
- `Deadline()` — when the context will be canceled, if a deadline is set.
- `Err()` — why `Done` closed, once it has (`nil` before that).
- `Value(key)` — the value associated with `key`, or `nil`.

```go
type CancelFunc func()
```

Idempotent — the first call to a `CancelFunc` does the real work; every subsequent call (even from other goroutines) is a no-op. It does *not* wait for the affected work to actually stop.

### Constructing a Context

`context.Background()` — a non-nil, empty context, never canceled, no deadline. The typical root for `main`, initialization, and top-level request handling.

`context.TODO()` — the same shape as `Background`, but signals "this should have a real context passed to it; that hasn't been wired up yet."

`context.WithValue(parent, key, val)` — derives a context carrying a key-value pair, inherited by everything derived from it. **Don't use this for anything a function actually depends on to behave correctly** — accept those as explicit parameters instead; reserve context values for genuinely cross-cutting, optional metadata (a request ID, for instance).

```go
ctx := context.Background()
ctx = context.WithValue(ctx, "processID", processID)
```

`context.WithCancel(parent)` — a derived context plus a `CancelFunc`. Always call `cancel` once the operation using the context is done, to release its resources — typically via `defer`, immediately after creation. Don't pass the `cancel` function around further than necessary.

`context.WithDeadline(parent, time.Time)` — cancels automatically at a specific point in time.

`context.WithTimeout(parent, duration)` — a thin wrapper: `WithDeadline(parent, time.Now().Add(timeout))`.

### Example: Detecting Request Cancellation

```go
func handleRequest(w http.ResponseWriter, req *http.Request) {
    ctx := req.Context()

    select {
    case <-time.After(5 * time.Second): // simulated work
        fmt.Fprintf(w, "Response from the server")
    case <-ctx.Done():
        fmt.Println("Error:", ctx.Err())
    }
}

func main() {
    http.HandleFunc("/request", handleRequest)
    http.ListenAndServe(":4000", nil)
}
```

If the client disconnects (`Ctrl+C` on `curl`, browser tab closed, etc.) before the 5 seconds elapse, `ctx.Done()` fires and the handler can stop the work early instead of finishing pointless computation for a response nobody will receive.

This project uses `context.Context` exactly this way throughout — see [`coding-standards.md §6`](coding-standards.md) (Concurrency) and [`§8`](coding-standards.md) (Graceful Shutdown) for how it's threaded through every handler and every service's shutdown path.

---

## Newer Standard Library Additions Worth Knowing

A few things that landed in the standard library after this course was originally written, relevant enough to flag on their own rather than as inline asides:

- **`log/slog`** (Go 1.21) — structured, leveled logging in the standard library, no third-party dependency required. This project uses it as its logging foundation — see [`coding-standards.md §4`](coding-standards.md).
- **`min`, `max`, `clear`** (Go 1.21) — built-in functions. `min`/`max` work on any ordered type; `clear` empties a map or zeroes a slice in place.
- **`slices` and `maps` packages** (Go 1.21) — generic helpers for the operations you'd otherwise write by hand constantly: `slices.Sort`, `slices.Contains`, `slices.Index`, `maps.Keys`, `maps.Clone`, and more.
- **`cmp` package** (Go 1.21) — `cmp.Ordered` (a ready-made constraint, no `golang.org/x/exp/constraints` dependency needed) and `cmp.Compare`.
- **Loop variable semantics** (Go 1.22) — covered inline in [Flow Control](#flow-control); the single most impactful change for anyone maintaining older Go code.
- **Range over integers** (Go 1.22) — `for i := range 10`.
- **Range-over-func iterators** (Go 1.23) — user-defined types can now be the target of a `for range` loop via a function matching `func(yield func(V) bool)` (or the two-value form). Lets you write your own lazy, `range`-able sequences without exposing a channel or building a full slice up front.
- **Generic type aliases** (Go 1.24) — `type Alias[T any] = SomeType[T]` — type aliases can now themselves be generic.

---

## Next Steps

You've now covered the fundamentals. From here:

- Apply it directly — this project's own [`setup-guide.md`](setup-guide.md), [`coding-standards.md`](coding-standards.md), and [`design-system.md`](design-system.md) are real, current Go code you're building against, not toy examples.
- [A Tour of Go](https://go.dev/tour/) — interactive, official.
- [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) — the community's own idiom guide.
- [Learn Go with Tests](https://quii.gitbook.io/learn-go-with-tests) — TDD-first, a good complement to this course's example-first style.

### References

- [The Go Programming Language](https://www.gopl.io/) (Donovan & Kernighan)
- [Official Go documentation](https://go.dev/doc/)
- [Official Go blog](https://go.dev/blog/)
- Original course: [karanpratapsingh.com — Learn Go: The Complete Course](https://www.karanpratapsingh.com/courses/go)
