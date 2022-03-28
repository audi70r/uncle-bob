# Uncle-bob
Golang clean architecture linter

![uncle bob](https://miro.medium.com/max/1400/1*nky1uZpFvkELX3RRC8tQSw.jpeg)

# Description

A golang linter based on Uncle Bob's clean code concepts.

The linter builds a hierarchical model from the project directory AST and assign a level for every 
package used by the project (with the exception of a standart golang library). The dependency 
levels are defined by the imports of a higher order package. If main package is considered level 0, then
all of its import will be 1. The subsequent imports of level 1 packages are respectively level 2 etc.

In plain mod, Uncle Bob will not allow same level imports.

In strict mod, Uncle Bob will only allow one level inward import (ex. level 0 can only import level 1 packages, level 1 can only import level 2 etc...)

Can by used in pipelines. If an issue is detected Uncle Bob will exit with status 1.

Linter works with go mod enabled

# Usage

Build the project
```bash
$ go build
```

Install the linter on Linux
```bash
$ go install
```

For running uncle bob, go to the project root directory, 
where the go.mod is located and simply run
```bash
$ uncle-bob
```

![uncle bob](uncle-bob-example.png)

Usage of uncle-bob:


To ignore test files
```bash
$ uncle-bob -ignore-tests
```

show detailed information about package imports
```bash
$ uncle-bob -package-imports=github.com/audi70r/uncle-bob/checker
``` 

do strict checking, allow only one level inward imports
```bash
$ uncle-bob -strict
``` 

# License
Do whatever you want with it, but don't disrespect Uncle Bob!
