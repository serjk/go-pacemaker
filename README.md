# Pacemaker

This library provides an API for connecting to and working with the
Pacemaker cluster manager, specifically with the cluster configuration
(the CIB), from the Go programming language.

It is not meant to be a complete API. The main use case is connecting
to the CIB, subscribing to updates and reading the XML.

**Note:** This API is under heavy development.

Current features:

* Connect and get CIB as an XML `[]byte` block

Major missing features:

* Decode CIB attributes and status section into a Go object structure
* Encode status section as JSON
* Decoding / encoding configuration section
* Writing changes back to the CIB


* Get CIB as JSON
* Get CibObjects as JSON
* Make changes
* Create CibObjects
* Get status of resources and nodes
* History information
* Meta information about agents etc.

## RUN UNIT TEST

To run the tests just run simple command :

    make test

## Usage

To include the library, import `github.com/krig/go-pacemaker`.

See `pacemaker_test.go` for usage examples.
