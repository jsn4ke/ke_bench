package kebench

// ConnectionPool is a generic connection pool implementation.
// It holds a ring buffer of connections and provides methods to get and push connections.
type ConnectionPool[In any] struct {
	New   func() (In, bool) // New is a function that creates a new connection.
	Close func(In)          // Close is a function that closes a connection.

	ring *Ring[In] // ring is the underlying ring buffer of connections.
}

func NewConnectionPool[In any](new func() (In, bool), close func(In), size int) *ConnectionPool[In] {
	return &ConnectionPool[In]{
		New:   new,
		Close: close,
		ring:  NewRing[In](size),
	}
}

// Get returns a connection from the pool.
// If there is an available connection in the ring buffer, it returns the connection and true.
// Otherwise, it calls the New function to create a new connection and returns it along with a boolean indicating if the connection was successfully created.
func (c *ConnectionPool[In]) Get() (In, bool) {
	in, ok := c.ring.Get()
	if ok {
		return in, ok
	}
	in, ok = c.New()
	return in, ok
}

// Push adds a connection to the pool.
// It first checks if there is an error or if the ring buffer is full.
// If either condition is true, it calls the Close function to close the connection and returns false.
// Otherwise, it pushes the connection to the ring buffer and returns true.
func (c *ConnectionPool[In]) Push(in In, err error) bool {
	if nil != err || !c.ring.Push(in) {
		c.Close(in)
		return false
	}
	return true
}
