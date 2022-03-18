package pubsub

import (
	"net/http"
	"testing"

	"github.com/onsi/gomega"
	"github.com/posener/wstest"
)

func TestSubscribe(t *testing.T) {
	var (
		s = &echoServer{}
		d = wstest.NewDialer(s)
	)

	gomega.RegisterTestingT(t)

	c, _, err := d.Dial("ws://example.org/ws", nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	_, p, err := c.ReadMessage()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(string(p)).To(gomega.ContainSubstring("server: new client"))

	c.WriteMessage(1, []byte("{\"op\": \"subscribe\", \"topic\": \"unit-testing\"}"))
	c.WriteMessage(1, []byte("{\"op\": \"publish\", \"topic\": \"unit-testing\", \"message\": \"this is an assertion\"}"))

	_, p, err = c.ReadMessage()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(string(p)).To(gomega.Equal("this is an assertion"))

	c.WriteMessage(1, []byte("{\"op\": \"publish\", \"topic\": \"unit-testing\", \"message\": \"this is an assertion2\"}"))
	_, p, err = c.ReadMessage()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(string(p)).To(gomega.Equal("this is an assertion2"))

	err = c.Close()
	if err != nil {
		panic(err)
	}

	<-s.Done
}

func TestUnsubscribe(t *testing.T) {
	var (
		s = &echoServer{}
		d = wstest.NewDialer(s)
	)

	gomega.RegisterTestingT(t)

	c, _, err := d.Dial("ws://example.org/ws", nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	_, p, err := c.ReadMessage()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(string(p)).To(gomega.ContainSubstring("server: new client"))

	c.WriteMessage(1, []byte("{\"op\": \"subscribe\", \"topic\": \"unit-testing\"}"))
	c.WriteMessage(1, []byte("{\"op\": \"subscribe\", \"topic\": \"unit-testing-2\"}"))
	c.WriteMessage(1, []byte("{\"op\": \"publish\", \"topic\": \"unit-testing\", \"message\": \"this is an assertion\"}"))

	_, p, err = c.ReadMessage()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(string(p)).To(gomega.Equal("this is an assertion"))

	c.WriteMessage(1, []byte("{\"op\": \"unsubscribe\", \"topic\": \"unit-testing\"}"))
	c.WriteMessage(1, []byte("{\"op\": \"publish\", \"topic\": \"unit-testing\", \"message\": \"this is an assertion2\"}"))
	c.WriteMessage(1, []byte("{\"op\": \"publish\", \"topic\": \"unit-testing-2\", \"message\": \"this is an assertion3\"}"))
	_, p, err = c.ReadMessage()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(string(p)).To(gomega.Equal("this is an assertion3"))

	err = c.Close()
	if err != nil {
		panic(err)
	}

	<-s.Done
}

type echoServer struct {
	Done chan struct{}
}

func (s *echoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Done = make(chan struct{})
	defer close(s.Done)

	WSHandler(w, r)
}
