package eventbus

import (
	"fmt"
	"testing"
	"time"
)

func TestEventBus_StartsAndStops(t *testing.T) {
	eb := New[int]()
	eb.Start()
	eb.Stop()

	select {
	case <-eb.done:
	case <-time.After(time.Second):
		t.Errorf("Expected EventBus to be stopped")
	}
}

func TestEventBus_PublishesData(t *testing.T) {
	eb := New[int]()
	eb.Start()
	defer eb.Stop()

	eb.Publish("test", 123)

	select {
	case n := <-eb.notifications:
		if n.topic != "test" || n.data != 123 {
			t.Errorf("Expected to receive published data")
		}
	case <-time.After(time.Second):
		t.Errorf("Expected to receive published data")
	}
}

func TestEventBus_SubscribesAndReceivesData(t *testing.T) {
	eb := New[int]()
	eb.Start()
	defer eb.Stop()

	dataReceived := false
	_, _ = eb.Subscribe("test", func(data int) error {
		if data != 123 {
			t.Errorf("Expected to receive published data")
		}
		dataReceived = true
		return nil
	})

	eb.Publish("test", 123)

	time.Sleep(250 * time.Millisecond)

	if !dataReceived {
		t.Errorf("Expected to receive published data")
	}
}

func TestEventBus_SubscribesWithExistingSubscriber(t *testing.T) {
	eb := New[int]()
	eb.Start()
	defer eb.Stop()

	cb := func(data int) error {
		return nil
	}

	_, err := eb.Subscribe("test", cb, Subscriber("subscriber"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, err = eb.Subscribe("test", cb, Subscriber("subscriber"))
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestEventBus_SubscribesWithReplace(t *testing.T) {
	eb := New[int]()
	eb.Start()
	defer eb.Stop()

	cb := func(data int) error {
		return nil
	}

	_, err := eb.Subscribe("test", cb, Subscriber("subscriber"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, err = eb.Subscribe("test", cb, Subscriber("subscriber"), Replace())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestEventBus_SubscribesWithError(t *testing.T) {
	eb := New[int]()
	eb.Start()
	defer eb.Stop()

	errReceived := false

	cb := func(data int) error {
		return fmt.Errorf("error")
	}
	errFn := func(err error) {
		errReceived = true
	}

	_, _ = eb.Subscribe("test", cb, Error(errFn))

	eb.Publish("test", 123)

	time.Sleep(150 * time.Millisecond)

	if !errReceived {
		t.Errorf("Expected to receive error")
	}
}

func TestEventBus_WithBufferSize(t *testing.T) {
	eb := New[int](BufferSize(1000))
	eb.Start()
	defer eb.Stop()

	if cap(eb.notifications) != 1000 {
		t.Errorf("Expected buffer size to be 100")
	}
}

func TestEventBus_MissingTopic(t *testing.T) {
	eb := New[int]()
	eb.Start()
	defer eb.Stop()

	_, err := eb.Subscribe("", func(data int) error {
		return nil
	})

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestEventBus_MissingHandler(t *testing.T) {
	eb := New[int]()
	eb.Start()
	defer eb.Stop()

	_, err := eb.Subscribe("test", nil)

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestEventBus_HandlerPanic(t *testing.T) {
	eb := New[int]()
	eb.Start()
	defer eb.Stop()

	var errRecovered bool

	_, _ = eb.Subscribe("test", func(data int) error {
		panic("panic")
	}, Error(func(err error) {
		errRecovered = true
	}))

	eb.Publish("test", 123)

	time.Sleep(150 * time.Millisecond)

	if !errRecovered {
		t.Errorf("Expected to recover from panic")
	}
}

func TestEventBus_UnsubscribeTopic(t *testing.T) {
	eb := New[int]()
	eb.Start()
	defer eb.Stop()

	var dataReceived bool

	_, _ = eb.Subscribe("test", func(data int) error {
		dataReceived = true
		return nil
	})

	eb.Unsubscribe(Topic("test"))

	eb.Publish("test", 123)

	time.Sleep(150 * time.Millisecond)

	if dataReceived {
		t.Errorf("Expected no data to be received")
	}
}

func TestEventBus_UnsubscribeSubscriber(t *testing.T) {
	eb := New[int]()
	eb.Start()
	defer eb.Stop()

	var data1Received bool
	var data2Received bool

	_, _ = eb.Subscribe("test1", func(data int) error {
		data1Received = true
		return nil
	}, Subscriber("subscriber"))

	_, _ = eb.Subscribe("test2", func(data int) error {
		data2Received = true
		return nil
	}, Subscriber("subscriber"))

	eb.Unsubscribe(Subscriber("subscriber"))

	eb.Publish("test1", 123)
	eb.Publish("test2", 123)

	time.Sleep(150 * time.Millisecond)

	if data1Received || data2Received {
		t.Errorf("Expected no data to be received")
	}
}

func TestEventBus_UnsubscribeTopicAndSubscriber(t *testing.T) {
	eb := New[int]()
	eb.Start()
	defer eb.Stop()

	var data1Received bool
	var data2Received bool

	_, _ = eb.Subscribe("test1", func(data int) error {
		data1Received = true
		return nil
	}, Subscriber("subscriber"))

	_, _ = eb.Subscribe("test2", func(data int) error {
		data2Received = true
		return nil
	}, Subscriber("subscriber"))

	eb.Unsubscribe(Topic("test1"), Subscriber("subscriber"))

	eb.Publish("test1", 123)
	eb.Publish("test2", 123)

	time.Sleep(150 * time.Millisecond)

	if data1Received {
		t.Errorf("Expected no data to be received")
	}

	if !data2Received {
		t.Errorf("Expected data to be received")
	}
}

func TestEventBus_UnsubscribeHandle(t *testing.T) {
	eb := New[int]()
	eb.Start()
	defer eb.Stop()

	var dataReceived bool

	sb, _ := eb.Subscribe("test", func(data int) error {
		dataReceived = true
		return nil
	})

	eb.Unsubscribe(Handle(sb.handle))

	eb.Publish("test", 123)

	if dataReceived {
		t.Errorf("Expected no data to be received")
	}
}
