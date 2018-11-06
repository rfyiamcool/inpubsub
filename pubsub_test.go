package inpubsub

import (
	"testing"
)

func TestNew(t *testing.T) {
	ps := NewPubSub()
	if ps == nil {
		t.Errorf("NewPubSub is nil")
	}

	got := len(ps.registries)
	if defaultPubSubTopicCapacity != got {
		t.Errorf("topic capacity is wrong. expected: %v, got: %v", defaultPubSubTopicCapacity, got)
	}

	for _, registory := range ps.registries {
		if "" != registory.topic {
			t.Errorf("topic word is wrong. expected: %v, got: %v", "", registory.topic)
		}
	}
}

func TestSubscribe(t *testing.T) {
	ps := NewPubSub()

	subscriber1 := ps.Subscribe("t1")
	subscriber2 := ps.Subscribe("t1")
	subscriber3 := ps.Subscribe("t2")

	ps.Publish("t1", "hi")
	ps.Publish("t2", "hello")

	got := <-subscriber1.Read()
	if "hi" != got {
		t.Errorf("message of subscriber1 is wrong. expected: %v, got: %v", "hi", got)
	}

	got = <-subscriber2.Read()
	if "hi" != got {
		t.Errorf("message of subscriber2 is wrong. expected: %v, got: %v", "hi", got)
	}

	got = <-subscriber3.Read()
	if "hello" != got {
		t.Errorf("message of subscribers is wrong. expected: %v, got: %v", "hello", got)
	}
}

func TestAddSubsrcibe(t *testing.T) {
	ps := NewPubSub()

	subscriber := ps.Subscribe("t1")
	ps.AddSubsrcibe("t2", subscriber)

	ps.Publish("t1", "hi")
	ps.Publish("t2", "hello")

	got := <-subscriber.Read()
	if "hi" != got {
		t.Errorf("message of subscribers is wrong. expected: %v, got: %v", "hi", got)
	}

	got = <-subscriber.Read()
	if "hello" != got {
		t.Errorf("message of subscribers is wrong. expected: %v, got: %v", "hello", got)
	}
}

func TestUnSubscribe(t *testing.T) {
	ps := NewPubSub()

	subscriber := ps.Subscribe("t1")
	ps.AddSubsrcibe("t2", subscriber)

	ps.UnSubscribe("t1", subscriber)

	ps.Publish("t1", "t1 - Hello World!!")
	ps.Publish("t2", "t2 - Hello World!!")

	got := <-subscriber.Read()

	if "t2 - Hello World!!" != got {
		t.Errorf("UnSubscribe is wrong. expected: %v, got: %v", "t2 - Hello World!!", got)
	}
}

func TestGenerateHash(t *testing.T) {
	tests := []struct {
		text     string
		expected uint32
	}{
		{
			"abc",
			uint32(1134309195),
		},
	}

	for i, test := range tests {
		got := generateHash(test.text)

		if test.expected != got {
			t.Errorf("tests[%d] - generateHash is wrong. expected: %v, got: %v", i, test.expected, got)
		}
	}
}
