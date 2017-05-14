package pntests

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	//pubnub "github.com/pubnub/go"
	pubnub ".."
)

var pnconfig *pubnub.PNConfiguration

func init() {
	pnconfig = pubnub.NewPNConfiguration()
	pnconfig.PublishKey = "my_pub_key"
	pnconfig.SubscribeKey = "my_sub_key"
	pnconfig.SecretKey = "my_secret_key"
	pnconfig.Origin = "localhost:3000"
	pnconfig.Secure = false
	pnconfig.ConnectionTimeout = 1
	pnconfig.NonSubscribeRequestTimeout = 2
}

func TestPublishContextTimeoutSync(t *testing.T) {
	assert := assert.New(t)
	seconds := 3
	timeout := time.Duration(seconds) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	shutdown := make(chan bool)
	go servePublish(pnconfig.NonSubscribeRequestTimeout+1, shutdown)

	pn := pubnub.NewPubNub(pnconfig)

	publish := pn.Publish()

	publish.Channel = "news"
	publish.Message = "hey"

	ok := make(chan interface{})
	err := make(chan error)

	go func() {
		resp, er := publish.ExecuteWithContext(ctx)
		if resp != nil {
			ok <- resp
		}

		if er != nil {
			err <- er
		}
	}()

	select {
	case resp := <-ok:
		assert.Fail("Received success instead of context deadline: %v", resp)
	case er := <-err:
		assert.Fail(fmt.Sprintf("Received error instead of context deadline: %v", er.Error()))
	case <-ctx.Done():
		assert.Equal(ctx.Err().Error(), "context deadline exceeded")
	case <-time.After(timeout + time.Duration(1)*time.Second):
		assert.Fail("Context cancellation doesn't work")
	}

	shutdown <- true
}

func TestContextCancelSync(t *testing.T) {
	assert := assert.New(t)
	ctx, cancel := context.WithCancel(context.Background())

	shutdown := make(chan bool)
	go servePublish(pnconfig.NonSubscribeRequestTimeout+1, shutdown)

	pn := pubnub.NewPubNub(pnconfig)
	publish := pn.Publish()

	publish.Channel = "news"
	publish.Message = "hey"

	ok := make(chan interface{})
	err := make(chan error)

	go func() {
		resp, er := publish.ExecuteWithContext(ctx)
		if resp != nil {
			ok <- resp
		}

		if er != nil {
			err <- er
		}
	}()

	cancel()

	select {
	case resp := <-ok:
		assert.Fail("Received success instead of context deadline: %v", resp)
	case er := <-err:
		assert.Fail(fmt.Sprintf("Received error instead of context deadline: %v", er.Error()))
	case <-ctx.Done():
		assert.Equal(ctx.Err().Error(), "context canceled")
	}

	shutdown <- true
}

func TestContextCancelAsync(t *testing.T) {
	done := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())

	go servePublish(pnconfig.NonSubscribeRequestTimeout+1, done)

	pn := pubnub.NewPubNub(pnconfig)
	publish := pn.Publish()

	ok := make(chan interface{})
	err := make(chan error)

	publish.Channel = "news"
	publish.Message = "hey"
	publish.SuccessChannel = ok
	publish.ErrorChannel = err

	go publish.ExecuteWithContext(ctx)

	cancel()

	select {
	case <-ctx.Done():
		assert.Equal(t, ctx.Err().Error(), "context canceled")
	}

	done <- true
}

func TestContextDeadlineAsync(t *testing.T) {
	done := make(chan bool)
	d := time.Now().Add(50 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), d)

	go servePublish(pnconfig.NonSubscribeRequestTimeout+1, done)

	pn := pubnub.NewPubNub(pnconfig)
	publish := pn.Publish()

	ok := make(chan interface{})
	err := make(chan error)

	publish.Channel = "news"
	publish.Message = "hey"
	publish.SuccessChannel = ok
	publish.ErrorChannel = err

	go publish.ExecuteWithContext(ctx)

	cancel()

	select {
	case <-ctx.Done():
		log.Println("Deadline exceeded")
	case resp := <-ok:
		log.Println(resp)
	case er := <-err:
		log.Fatal(er)
	}

	done <- true
}

func AestPublishTimeoutAsync(t *testing.T) {
	done := make(chan bool)
	go servePublish(pnconfig.NonSubscribeRequestTimeout+1, done)

	pn := pubnub.NewPubNub(pnconfig)

	publish := pn.Publish()

	publish.Channel = "news"
	publish.Message = "hey"
	// Async
	ok := make(chan interface{})
	err := make(chan error)

	publish.SuccessChannel = ok
	publish.ErrorChannel = err

	go publish.Execute()
	fmt.Println("async gone..")

	select {
	case resp := <-ok:
		log.Println(resp)
	case er := <-err:
		log.Println(er)
	}

	done <- true
}

func makeResponseRoot(hangSeconds int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(hangSeconds) * time.Second)
		fmt.Fprint(w, "[1, \"Sent\", 123]")
	}
}

func servePublish(hangSeconds int, done chan bool) {
	r := mux.NewRouter()
	r.HandleFunc("/publish/{pubKey}/{subKey}/0/{channel}/0/{msg}",
		makeResponseRoot(hangSeconds))

	s := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}

	go s.ListenAndServe()

	<-done
	log.Println("closing server")
	s.Close()
}