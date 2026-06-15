package global

import (
	"sync"

	sio "github.com/doquangtan/socketio/v4"
)

var (
	io *sio.Io
	n  sync.Once
)

func GetSocketIO() *sio.Io {
	n.Do(func() {
		io = sio.New()
	})
	return io
}
