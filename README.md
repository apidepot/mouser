# mouser

Go-based client for accessing the Mouser API.

[![GoDoc][godoc badge]][godoc link]
[![Go Report Card][report badge]][report card]
[![License Badge][license badge]][LICENSE.txt]

## Overview

The [mouser][] module provides a Go-based client for the [Mouser
API][m-api]. To access the [Mouser API][m-api] your client application must
be registered with DigiKey and to make production calls to the API, developers
must be a member of an organization. To learn more see the [Mouser API
Solutions][m-solutions].

## Installation

```bash
$ go get github.com/apidepot/digikey
```

## Examples

Examples are available at <https://github.com/apidepot/digikey-examples/>.

## Implementation Status

This library is currently in alpha status and is changing frequently. Not
everything is implemented, including a list of what is implemented.

## Contributing

Contributions are welcome! To contribute please:

1. Fork the repository
2. Create a feature branch
3. Code
4. Submit a [pull request][]

### Testing

Instead of using [GNU Make][make], this project uses [Just][] as its
task/command runner.

Prior to submitting a [pull request][], please run:

```bash
$ just check    # formats, vets, and unit tests the code
$ just lint     # lints code using staticcheck
```

To update and view the test coverage report:

```bash
$ just cover
```

#### Integration Testing

To perform the integration tests run:

```bash
$ just int
```

## License

[mouser][] is released under the MIT license. Please see the
[LICENSE.txt][] file for more information.

[godoc badge]: https://godoc.org/github.com/apidepot/mouser?status.svg
[godoc link]: https://godoc.org/github.com/apidepot/mouser
[just]: https://just.systems/
[LICENSE.txt]: https://github.com/apidepot/mouser/blob/master/LICENSE.txt
[license badge]: https://img.shields.io/badge/license-MIT-blue.svg
[m-api]: https://www.mouser.com/api-hub/
[m-solutions]: https://www.mouser.com/api-solutions/
[make]: https://www.gnu.org/software/make/
[mouser]: https://github.com/apidepot/mouser
[pull request]: https://help.github.com/articles/using-pull-requests
[report badge]: https://goreportcard.com/badge/github.com/apidepot/mouser
[report card]: https://goreportcard.com/report/github.com/apidepot/digikey
