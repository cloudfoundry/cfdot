# cfdot

`cfdot` is the CF Diego Operator Toolkit, a CLI tool designed to interact with
Diego components.

At present, its commands focus on the Diego BBS API, the main public interface
to a Diego deployment.


## Current Commands

```bash
$ cfdot --help
A command-line tool to interact with a Cloud Foundry Diego deployment

Usage:
  cfdot [command]

Available Commands:
  actual-lrp-groups            List actual LRP groups
  actual-lrp-groups-for-guid   List actual LRP groups for a process guid
  cancel-task                  Cancel task
  cell                         Show the specified cell presence
  cell-state                   Show the specified cell state
  cells                        List registered cell presences
  claim-lock                   Claim Locket lock
  claim-presence               Claim Locket presence
  create-desired-lrp           Create a desired LRP
  create-task                  Create a Task
  delete-desired-lrp           Delete a desired LRP
  delete-task                  Delete a Task
  desired-lrp                  Show the specified desired LRP
  desired-lrp-scheduling-infos List desired LRP scheduling infos
  desired-lrps                 List desired LRPs
  domains                      List domains
  help                         Get help on [command]
  locks                        List Locket locks
  lrp-events                   Subscribe to BBS LRP events
  presences                    List Locket presences
  release-lock                 Release Locket lock
  retire-actual-lrp            Retire actual LRP by index and process guid
  set-domain                   Set domain
  task                         Display task
  task-events                  Subscribe to BBS Task events
  tasks                        List tasks in BBS
  update-desired-lrp           Update a desired LRP

Flags:
  -h, --help   help for cfdot

Use "cfdot [command] --help" for more information about a command.

```

## Running from a BOSH-deployed VM

`cfdot` is most useful in the context of a running Diego deployment.  If you
use the [`generate-deployment-manifest`](https://github.com/cloudfoundry/diego-release/blob/master/scripts/generate-deployment-manifest)
script in diego-release to generate your Diego manifest, `cfdot` is already
available on the BOSH-deployed Diego VMs. To use it:

```bash
bosh ssh <DIEGO_JOB>/<INDEX>
cfdot
```

The `cfdot` pre-start script installs the `setup` script into `/etc/profile.d`.
This `setup` script does 3 things:

- Exports environment variables to target the BBS API in the deployment.
- Puts the `cfdot` binary on the `PATH`.
- Puts a `jq` binary on the `PATH`.

## Basic Examples

```bash
# count the total number of desired instances
$ cfdot desired-lrp-scheduling-infos | jq '.instances' | jq -s 'add'
568

# show instance counts by state
$ cfdot actual-lrp-groups | jq '.instance, .evacuating | values' | jq -s -r 'group_by(.state)[] | .[0].state + ": " + (length | tostring)'
CRASHED: 36
RUNNING: 531
UNCLAIMED: 1
```

## Building from Source

`cfdot` requires the [Diego BBS client library](https://github.com/cloudfoundry/bbs).
If you have already cloned [diego-release](https://github.com/cloudfoundry/diego-release),
you can run the following commands using that diego-release directory as your
GOPATH.  Alternatively, run these commands with any other GOPATH and `go get`
will automatically fetch the latest BBS code from diego-release.

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

## Design Tenets

- Execution is stateless: configuration is specified either as flags or as environment variables.
- Conform to UNIX conventions of successful output on stdout and error messages on stderr.
- For BBS API commands, output is a stream of JSON values, one per line, optimal for processing with `jq` and suitable for processing with `bash` and other line-based UNIX utilities.
