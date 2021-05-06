package app

import (
	"io/ioutil"
	"net"
	"os"
	"runtime"

	"github.com/mattermost/mattermost-server/v5/model"
)


type MyStruct struct {
	s *Server
}

func NewGlue(s *Server) *MyStruct {
	return &MyStruct{s}
}

func (m *MyStruct) MyTest(channelID string, out *model.GetChannelRPCResponse) error {
	a := New(ServerConnector(m.s))

	ch, err := a.GetChannel(channelID)
	if err != nil {
		return err
	}
	*out = model.GetChannelRPCResponse{
		Channel: ch,
	}
	return nil
}


func serverListener() (net.Listener, error) {
	if runtime.GOOS == "windows" {
		return serverListener_tcp()
	}

	return serverListener_unix()
}

func serverListener_tcp() (net.Listener, error) {
	listener, err := net.Listen("tcp", "localhost:")
	if err != nil {
		return nil, err
	}

	return listener, nil
}

func serverListener_unix() (net.Listener, error) {
	tf, err := ioutil.TempFile("", "suite")
	if err != nil {
		return nil, err
	}
	path := tf.Name()

	// Close the file and remove it because it has to not exist for
	// the domain socket.
	if err := tf.Close(); err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil {
		return nil, err
	}

	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	// Wrap the listener in rmListener so that the Unix domain socket file
	// is removed on close.
	return &rmListener{
		Listener: l,
		Path:     path,
	}, nil
}

// rmListener is an implementation of net.Listener that forwards most
// calls to the listener but also removes a file as part of the close. We
// use this to cleanup the unix domain socket on close.
type rmListener struct {
	net.Listener
	Path string
}

func (l *rmListener) Close() error {
	// Close the listener itself
	l.Listener.Close()

	// Remove the file
	return os.Remove(l.Path)
}
