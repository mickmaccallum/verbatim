package states

import "fmt"

var ErrCaptionersStillConnected = fmt.Errorf("There are still captioners connected on this port. Disconnect them before changing it.")
