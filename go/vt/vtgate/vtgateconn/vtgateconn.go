// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vtgateconn

import (
	"flag"
	"time"

	log "github.com/golang/glog"
	mproto "github.com/youtube/vitess/go/mysql/proto"
	tproto "github.com/youtube/vitess/go/vt/tabletserver/proto"
	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/vtgate/proto"
	"golang.org/x/net/context"
)

const (
	// GoRPCProtocol is a vtgate protocol based on go rpc
	GoRPCProtocol = "gorpc"
)

var (
	vtgateProtocol = flag.String("vtgate_protocol", GoRPCProtocol, "how to talk to vtgate")
)

// ServerError represents an error that was returned from
// a vtgate server.
type ServerError struct {
	Code int
	Err  string
}

func (e *ServerError) Error() string { return e.Err }

// OperationalError represents an error due to a failure to
// communicate with vtgate.
type OperationalError string

func (e OperationalError) Error() string { return string(e) }

// DialerFunc represents a function that will return a VTGateConn object that can communicate with a VTGate.
type DialerFunc func(ctx context.Context, address string, timeout time.Duration) (VTGateConn, error)

// VTGateConn defines the interface for a vtgate client.
// It can be used concurrently across goroutines.
type VTGateConn interface {
	// Execute executes a non-streaming query on vtgate.
	Execute(ctx context.Context, query string, bindVars map[string]interface{}, tabletType topo.TabletType) (*mproto.QueryResult, error)
	// ExecuteShard executes a non-streaming query for multiple shards on vtgate.
	ExecuteShard(ctx context.Context, query string, keyspace string, shards []string, bindVars map[string]interface{}, tabletType topo.TabletType) (*mproto.QueryResult, error)

	// StreamExecute executes a streaming query on vtgate. It returns a channel, ErrFunc and error.
	// If error is non-nil, it means that the StreamExecute failed to send the request. Otherwise,
	// you can pull values from the channel till it's closed. Following this, you can call ErrFunc
	// to see if the stream ended normally or due to a failure.
	StreamExecute(ctx context.Context, query string, bindVars map[string]interface{}, tabletType topo.TabletType) (<-chan *mproto.QueryResult, ErrFunc)

	// Begin starts a transaction and returns a VTGateTX.
	Begin(ctx context.Context) (VTGateTx, error)

	// Close must be called for releasing resources.
	Close()

	// SplitQuery splits a query into equally sized smaller queries by
	// appending primary key range clauses to the original query
	SplitQuery(ctx context.Context, keyspace string, query tproto.BoundQuery, splitCount int) ([]proto.SplitQueryPart, error)
}

// VTGateTx defines the interface for the transaction object created by Begin.
// It should not be concurrently used across goroutines.
type VTGateTx interface {
	// Execute executes a query on vtgate within the current transaction.
	Execute(ctx context.Context, query string, bindVars map[string]interface{}, tabletType topo.TabletType) (*mproto.QueryResult, error)
	// ExecuteShard executes a query for multiple shards on vtgate within the current transaction.
	ExecuteShard(ctx context.Context, query string, keyspace string, shards []string, bindVars map[string]interface{}, tabletType topo.TabletType) (*mproto.QueryResult, error)

	// Commit commits the current transaction.
	Commit(ctx context.Context) error
	// Rollback rolls back the current transaction.
	Rollback(ctx context.Context) error
}

// ErrFunc is used to check for streaming errors.
type ErrFunc func() error

var dialers = make(map[string]DialerFunc)

// RegisterDialer is meant to be used by Dialer implementations
// to self register.
func RegisterDialer(name string, dialer DialerFunc) {
	if _, ok := dialers[name]; ok {
		log.Warningf("Dialer %s already exists", name)
		return
	}
	dialers[name] = dialer
}

// GetDialer returns the dialer to use, described by the command line flag
func GetDialer() DialerFunc {
	return GetDialerWithProtocol(*vtgateProtocol)
}

// GetDialerWithProtocol returns the dialer to use, described by the given protocol
func GetDialerWithProtocol(protocol string) DialerFunc {
	td, ok := dialers[protocol]
	if !ok {
		log.Warningf("No dialer registered for VTGate protocol %s", protocol)
		return nil
	}
	return td
}
