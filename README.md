# Pacemaker

This library provides an API for connecting to and working with the
Pacemaker cluster manager, specifically with the cluster configuration
(the CIB), from the Go programming language.

**Note:** This API is under development.

Current features:

* 	CreateObjInSection
* 	UpdateObjInSection
*   ReplaceObjInSection
*   DeleteObjInSection
  
*   Query
*   QueryXPathNoChildren
*   QueryXPath
*   Version
  
*   GetLocalNodeName
*   GetNodesInfo
*   GetNodeIp

For more information have a look into cib.go

Major missing features:

* History information
* Meta information about agents etc.

## RUN UNIT TESTS

To run the tests just run simple command :

    make test

## Usage

To include the library, import `github.com/serjk/go-pacemaker`.

See `./impl/pacemaker_test.go` for usage examples.
