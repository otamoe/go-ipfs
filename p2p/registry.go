package p2p

import (
	"fmt"
	"io"

	ma "gx/ipfs/QmcyqRMCAXVtYPS4DiBrA7sezL9rRGfW8Ctx7cywL4TXJj/go-multiaddr"
	peer "gx/ipfs/QmdS9KpbDyPrieswibZhkod1oXqRwZJrUPzxCofAMWpFGq/go-libp2p-peer"
)

// ListenerInfo holds information on a p2p listener.
type ListenerInfo struct {
	// Application protocol identifier.
	Protocol string

	// Node identity
	Identity peer.ID

	// Local protocol stream address.
	Address ma.Multiaddr

	// Local protocol stream listener.
	Closer io.Closer

	// Flag indicating whether we're still accepting incoming connections, or
	// whether this application listener has been shutdown.
	Running bool

	Registry *ListenerRegistry
}

// Close closes the listener. Does not affect child streams
func (c *ListenerInfo) Close() error {
	c.Closer.Close()
	err := c.Registry.Deregister(c.Protocol)
	return err
}

// ListenerRegistry is a collection of local application protocol listeners.
type ListenerRegistry struct {
	Listeners []*ListenerInfo
}

// Register registers listenerInfo2 in this registry
func (c *ListenerRegistry) Register(listenerInfo *ListenerInfo) {
	c.Listeners = append(c.Listeners, listenerInfo)
}

// Deregister removes p2p listener from this registry
func (c *ListenerRegistry) Deregister(proto string) error {
	foundAt := -1
	for i, a := range c.Listeners {
		if a.Protocol == proto {
			foundAt = i
			break
		}
	}

	if foundAt != -1 {
		c.Listeners = append(c.Listeners[:foundAt], c.Listeners[foundAt+1:]...)
		return nil
	}

	return fmt.Errorf("failed to deregister proto %s", proto)
}

// StreamInfo holds information on active incoming and outgoing p2p streams.
type StreamInfo struct {
	HandlerID uint64

	Protocol string

	LocalPeer peer.ID
	LocalAddr ma.Multiaddr

	RemotePeer peer.ID
	RemoteAddr ma.Multiaddr

	Local  io.ReadWriteCloser
	Remote io.ReadWriteCloser

	Registry *StreamRegistry
}

// Close closes stream endpoints and deregisters it
func (s *StreamInfo) Close() error {
	s.Local.Close()
	s.Remote.Close()
	s.Registry.Deregister(s.HandlerID)
	return nil
}

func (s *StreamInfo) startStreaming() {
	go func() {
		io.Copy(s.Local, s.Remote)
		s.Close()
	}()

	go func() {
		io.Copy(s.Remote, s.Local)
		s.Close()
	}()
}

// StreamRegistry is a collection of active incoming and outgoing protocol app streams.
type StreamRegistry struct {
	Streams []*StreamInfo

	nextID uint64
}

// Register registers a stream to the registry
func (c *StreamRegistry) Register(streamInfo *StreamInfo) {
	streamInfo.HandlerID = c.nextID
	c.Streams = append(c.Streams, streamInfo)
	c.nextID++
}

// Deregister deregisters stream from the registry
func (c *StreamRegistry) Deregister(handlerID uint64) {
	foundAt := -1
	for i, s := range c.Streams {
		if s.HandlerID == handlerID {
			foundAt = i
			break
		}
	}

	if foundAt != -1 {
		c.Streams = append(c.Streams[:foundAt], c.Streams[foundAt+1:]...)
	}
}