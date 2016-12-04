package states

import "fmt"

// ErrCaptionersStillConnected indicates that there are still captioners connected on a given port.
var ErrCaptionersStillConnected = fmt.Errorf("There are still captioners connected on this port. Disconnect them before changing it.")
