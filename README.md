# Process List Library for Go

go-ps is a library for Go that implements OS-specific APIs to list and
manipulate processes in a platform-safe way. The library can find and
list processes on Linux, Mac OS X, and Windows.

If you're new to Go, this library has a good amount of advanced Go educational
value as well. It uses some advanced features of Go: build tags, accessing
DLL methods for Windows, cgo for Darwin, etc.

## Installation

Install using standard `go get`:

```
$ go get github.com/mitchellh/go-ps
...
```

## TODO

Want to contribute? Here is a short TODO list of things that aren't
implemented for this library that would be nice:

  * FreeBSD support
  * Plan9 support
  * Eliminate the need for cgo with Darwin
