# gitodo

## Overview

[Motivation blog post](https://www.drazengolic.com/blog/committing-upfront/)

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

Currently it's only possible to install the application via source.

Prerequisites:

1. [Go](https://go.dev/dl/) 1.23.4 or newer
2. GCC
    1. MacOS: XCode, or type `xcode-select --install` in Terminal if you don't want the full app
    2. Linux: install `build-essential` package or equivalent
    3. Windows: install MSYS2 package in your PATH, or if you're using Scoop: `scoop install gcc`
    
After the prerequisites are installed, execute the following command:

`go install github.com/drazengolic/gitodo@latest`

## Usage

Type `gitodo help` or `gitodo help <command>` to see usage. Or view it [online](https://www.drazengolic.com/gitodo/).

## Licence

gitodo is released under the Apache 2.0 license. See [LICENCE](LICENCE)

copyright © Dražen Golić
