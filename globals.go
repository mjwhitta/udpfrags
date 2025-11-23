package udpfrags

// Version is the package version.
const Version string = "1.4.2"

var (
	// QueueSize is the size of the receiving queue.
	QueueSize uint64 = 1024

	// Size of each fragment when receiving/sending.
	bufSize uint64 = 1024 // 1KB

	// Size of the fragment header [fragNum/fragsTotal]
	fragHdrSize int = 16
)
