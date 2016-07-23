# cfdot

## Building from source

`cfdot` requires the BBS client library from diego-release, so if you already a cloned diego-release,
you can run the following commands using that diego-release directory as your GOPATH. Alternatively, run
these commands with any other GOPATH and `go get` will automatically fetch the latest BBS code from diego-release.

```bash
# Get cfdot and required dependencies
go get code.cloudfoundry.org/cfdot
cd src/code.cloudfoundry.org/cfdot

# Build for Linux
GOOS=linux go build .

# Build for Mac
GOOS=darwin go build .

# Build for Windows
GOOS=windows go build .
```
