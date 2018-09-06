package faye_test

import (
	"fmt"
	"github.com/thesyncim/faye"
	"github.com/thesyncim/faye/extensions"
	"github.com/thesyncim/faye/message"
	"log"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

type cancelFn func()

func setup(t *testing.T) (cancelFn, error) {
	cmd := exec.Command("npm", "start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	var cancel = func() {
		exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprint(cmd.Process.Pid)).Run()
		log.Println("canceled")
	}
	go func() {
		select {
		case <-time.After(time.Second * 30):
			cancel()
			t.Fatal("failed")
			os.Exit(1)

		}
	}()

	return cancel, nil
}

func TestServerSubscribeAndPublish(t *testing.T) {
	shutdown, err := setup(t)
	if err != nil {
		t.Fatal(err)
	}
	defer shutdown()

	debug := extensions.NewDebugExtension(os.Stdout)

	client, err := fayec.NewClient("ws://localhost:8000/faye", fayec.WithExtension(debug.InExtension, debug.OutExtension))
	if err != nil {
		t.Fatal(err)
	}

	client.OnPublishResponse("/test", func(message *message.Message) {
		if !message.Successful {
			t.Fatalf("failed to send message with id %s", message.Id)
		}
	})
	var done sync.WaitGroup
	done.Add(10)
	var delivered int
	go func() {
		client.Subscribe("/test", func(data message.Data) {
			if data != "hello world" {
				t.Fatalf("expecting: `hello world` got : %s", data)
			}
			delivered++
			done.Done()
		})
	}()

	//give some time for setup
	time.Sleep(time.Second)
	for i := 0; i < 10; i++ {
		id, err := client.Publish("/test", "hello world")
		if err != nil {
			t.Fatal(err)
		}
		log.Println(id, i)
	}

	done.Wait()
	err = client.Unsubscribe("/test")
	if err != nil {
		t.Fatal(err)
	}
	if delivered != 10 {
		t.Fatal("message received after client unsubscribe")
	}
	log.Println("complete")

}

func TestSubscribeUnauthorizedChannel(t *testing.T) {
	shutdown, err := setup(t)
	if err != nil {
		t.Fatal(err)
	}
	defer shutdown()

	debug := extensions.NewDebugExtension(os.Stdout)

	client, err := fayec.NewClient("ws://localhost:8000/faye", fayec.WithExtension(debug.InExtension, debug.OutExtension))
	if err != nil {
		t.Fatal(err)
	}

	err = client.Subscribe("/unauthorized", func(data message.Data) {
		t.Fatal("received message on unauthorized channel")
	})
	if err == nil {
		t.Fatal("subscribed to an unauthorized channel")
	}

}
