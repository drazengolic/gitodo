# Welcome to gitodo docs

To see what is gitodo and what can it do, [check here](gitodo.md).

## Installation

Currently it's only possible to install the application via source.

### Prerequisites:

1. [Go](https://go.dev/dl/) 1.23.4 or newer
2. A C compiler
    1. MacOS: XCode, or type `xcode-select --install` in Terminal if you don't want the full app
    2. Linux: install `build-essential` package or equivalent
    3. Windows: MSYS2 package ([instructions](https://medium.com/@freschiandrea86/how-to-use-go-and-cgo-in-windows-9014768d0f73))
    
After the prerequisites are installed, execute the following command:

`go install github.com/drazengolic/gitodo@latest`

The application is tested on macOS and Ubuntu Linux. Windows version is not tested yet, and may have issues.