# cfdot

## Building from source

To get the compatible version of the BBS client, it is best to follow the `GOPATH` structure of [diego-release](code.cloudfoundry.org/diego-release). If you already have a working diego-release directory, you can `cd` into there and skip the "Get diego-release" step. Run the following:

```bash
# Get diego-release
git clone https://github.com/cloudfoundry/diego-release
cd diego-release
export GOPATH=$PWD
export PATH=$GOPATH/bin:$PATH
./scripts/update

# Get cfdot and required dependency
go get code.cloudfoundry.org/cfdot
go get github.com/jessevdk/go-flags
cd src/code.cloudfoundry.org/cfdot

# Build for Linux
GOOS=linux go build ./cmd/cfdot/

# Build for Mac
GOOS=darwin go build ./cmd/cfdot/

# Build for Windows
GOOS=windows go build ./cmd/cfdot/
```
