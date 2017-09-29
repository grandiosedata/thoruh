# thoruh

[![Travis](https://img.shields.io/travis/jonathanmarvens/thoruh.svg?style=flat-square)](https://travis-ci.org/jonathanmarvens/thoruh)

A __Go__ package for parsing GNU-style (and DOS-style) command-line options (*inspired by the awesome __argtable2__ __C__ library*).

## Installation

```sh
go get -u -v github.com/grandiosedata/thoruh
```

## Usage

```go
package main

import (
	// "fmt"
	// "os"

	// "github.com/grandiosedata/thoruh"
)


func main() {
	// --foo -x123 /x456 /x:"789" /x / /x:"/"
	// fmt.Printf("%#v\n", os.Args) // []string{"--foo", "-x123", "/x456", "/x:789", "/x", "/", "/x:/"}
}

// TODO: Complete usage example …
```

## Author

__Jonathan Barronville__ <[jonathan@belairlabs.com](mailto:jonathan@belairlabs.com)>

## License

```
Copyright 2017 Jonathan Barronville <jonathan@belairlabs.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
