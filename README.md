# gitodo

## Overview

gitodo is a to-do list companion for git projects that ties to-do items to git 
repositories and branches without storing any files in the actual repositories.

A minimalist tool that helps the busy developers to:
 - keep track of what they've done and what they need
   to do per branch
 - add ideas in the queue for later
 - make stashing and popping of changes easier
 - craft commit messages based on the work done
 - prepare changelists
 - track time
 - view reports
 
**Note:** The application is still under development, but the major features have been completed and it is considered to be in the usable state.
 
## Installation

Currently it's only possible to install the application via source. Packages for all major platforms are planned.

Prerequisites:

1. [Go](https://go.dev/dl/) 1.23.4 or newer
2. A C compiler
    1. MacOS: XCode, or type `xcode-select --install` in Terminal if you don't want the full app
    2. Linux: install `build-essential` package or equivalent
    3. Windows: MSYS2 package ([instructions](https://medium.com/@freschiandrea86/how-to-use-go-and-cgo-in-windows-9014768d0f73))
    
After the prerequisites are installed, execute the following command:

`go install github.com/drazengolic/gitodo@master`

## Usage

Type `gitodo help` or `gitodo help <command>` to see usage.

## Licence

gitodo is released under the Apache 2.0 license. See [LICENCE](LICENCE)

copyright ©️ Dražen Golić
