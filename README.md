# catfact

A command line for catfact.

`catfact` is a single pure-Go binary. It reads public catfact data
over plain HTTPS, shapes it into clean records, and prints output that pipes
into the rest of your tools. No API key, nothing to run alongside it.

The same package is also a [resource-URI driver](#use-it-as-a-resource-uri-driver),
so a host program like [ant](https://github.com/tamnd/ant) can address
catfact as `catfact://` URIs.

## Install

```bash
go install github.com/tamnd/catfact-cli/cmd/catfact@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/catfact-cli/releases), or run
the container image:

```bash
docker run --rm ghcr.io/tamnd/catfact:latest --help
```

## Usage

```bash
catfact page <path>                      # fetch one page as a record
catfact page <path> -o json              # as JSON, ready for jq
catfact page <path> --template '{{.Body}}'  # just the readable body text
catfact links <path>                     # the pages it links to, one per line
catfact --help                           # the whole command tree
```

Every command shares one output contract: `-o table|json|jsonl|csv|tsv|url|raw`,
`--fields` to pick columns, `--template` for a custom line, and `-n` to limit.
The default adapts to where output goes (a table on a terminal, JSONL in a
pipe), so the same command reads well by hand and parses cleanly downstream.

This is a fresh scaffold. It ships one example resource type, `page`, wired end
to end. Model the real catfact records in `catfact/` and declare their
operations in `catfact/domain.go`; each one becomes a command, an HTTP
route, and an MCP tool at once.

## Serve it

The same operations are available over HTTP and as an MCP tool set for agents,
with no extra code:

```bash
catfact serve --addr :7777    # GET /v1/page/<path>  returns NDJSON
catfact mcp                   # speak MCP over stdio
```

## Use it as a resource-URI driver

`catfact` registers a `catfact` domain the way a program registers a
database driver with `database/sql`. A host enables it with one blank import:

```go
import _ "github.com/tamnd/catfact-cli/catfact"
```

Then [ant](https://github.com/tamnd/ant) (or any program that links the package)
dereferences `catfact://` URIs without knowing anything about catfact:

```bash
ant get catfact://page/<path>   # fetch the record
ant cat catfact://page/<path>   # just the body text
ant ls  catfact://page/<path>   # the pages it links to, each addressable
ant url catfact://page/<path>   # the live https URL
```

## Development

```
cmd/catfact/   thin main: hands cli.NewApp to kit.Run
cli/                 assembles the kit App from the catfact domain
catfact/                the library: HTTP client, data models, and domain.go (the driver)
docs/                tago documentation site
```

```bash
make build      # ./bin/catfact
make test       # go test ./...
make vet        # go vet ./...
```

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds the
archives, Linux packages, the multi-arch GHCR image, checksums, SBOMs, and a
cosign signature:

```bash
git tag v0.1.0
git push --tags
```

The Homebrew and Scoop steps self-disable until their tokens exist, so the first
release works with no extra secrets.

## License

Apache-2.0. See [LICENSE](LICENSE).
