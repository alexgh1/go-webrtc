package data

import (
	// "fmt"
	"reflect"
	"testing"
	"time"
)

var c *Channel

func checkEnum(t *testing.T, desc string, enum int, expected int) {
	if enum != expected {
		t.Error("Mismatched Enum Value -", desc,
			"\nwas:", enum,
			"\nexpected:", expected)
	}
}

func TestDataStateEnums(t *testing.T) {
	checkEnum(t, "DataStateConnecting",
		int(DataStateConnecting), _cgoDataStateConnecting)
	checkEnum(t, "DataStateOpen",
		int(DataStateOpen), _cgoDataStateOpen)
	checkEnum(t, "DataStateClosing",
		int(DataStateClosing), _cgoDataStateClosing)
	checkEnum(t, "DataStateClosed",
		int(DataStateClosed), _cgoDataStateClosed)
}

func TestNewChannel(t *testing.T) {
	c = NewChannel(cgoFakeDataChannel())
	if nil == c {
		t.Fatal("Could not create NewChannel")
	}
}

// There's not a good way to create a DataChannel without first having
// an available PeerConnection object with a valid session, but that's part of
// the outer package, so testing the attributes is less useful here.

func TestChannelLabel(t *testing.T) {
	if "fake" != c.Label() {
		t.Error()
	}
}

func TestChannelOrdered(t *testing.T) {
	if false != c.Ordered() {
		t.Error()
	}
}

func TestChannelReadyState(t *testing.T) {
	if DataStateClosed != c.ReadyState() {
		t.Error()
	}
}

func TestOnMessageCallback(t *testing.T) {
	success := make(chan []byte, 1)
	c.OnMessage = func(msg []byte) {
		success <- msg
	}
	bytes := []byte("somenumberofbytesinhere")
	size := len(bytes)
	cgoFakeMessage(c, bytes, size)
	select {
	case data := <-success:
		if !reflect.DeepEqual(data, bytes) {
			t.Error("Unexpected bytes: ", data)
		}
	case <-time.After(time.Second * 1):
		t.Fatal("Timed out.")
	}
}

func TestStateChangeCallbacks(t *testing.T) {
	opened := make(chan int, 1)
	closed := make(chan int, 1)
	c.OnOpen = func() {
		opened <- 1
	}
	c.OnClose = func() {
		closed <- 1
	}

	cgoFakeStateChange(c, DataStateOpen)
	select {
	case <-opened:
		if DataStateOpen != c.ReadyState() {
			t.Error("Unexpected state: ", c.ReadyState())
		}
	case <-time.After(time.Second * 1):
		t.Fatal("Timed out when waiting for Open.")
	}

	cgoFakeStateChange(c, DataStateClosed)
	select {
	case <-closed:
		if DataStateClosed != c.ReadyState() {
			t.Error("Unexpected state: ", c.ReadyState())
		}
	case <-time.After(time.Second * 1):
		t.Fatal("Timed out when waiting for Closed.")
	}
	// Set to open for the next tests.
	cgoFakeStateChange(c, DataStateOpen)
}

func TestOnBufferedAmountLowCallback(t *testing.T) {
	success := make(chan int, 1)
	c.BufferedAmountLowThreshold = 100
	c.OnBufferedAmountLow = func() {
		success <- 1
	}
	cgoFakeBufferAmount(c, 90)
	select {
	case <-success:
	case <-time.After(time.Second * 1):
		t.Fatal("Timed out.")
	}
}

func TestSend(t *testing.T) {
	messages := make(chan []byte, 1)
	data := []byte("some data to send")
	c.OnMessage = func(msg []byte) {
		messages <- msg
	}
	c.Send(data)
	select {
	case recv := <-messages:
		if !reflect.DeepEqual(recv, data) {
			t.Error("Unexpected bytes: ", recv)
		}
	case <-time.After(time.Second * 1):
		t.Fatal("Timed out.")
	}
}

func TestCloseChannel(t *testing.T) {
	closed := make(chan int, 1)
	c.OnClose = func() {
		closed <- 1
	}
	c.Close()
	select {
	case <-closed:
		if DataStateClosed != c.ReadyState() {
			t.Error("Unexpected state: ", c.ReadyState())
		}
	case <-time.After(time.Second * 1):
		t.Fatal("Timed out during close..")
	}
}
