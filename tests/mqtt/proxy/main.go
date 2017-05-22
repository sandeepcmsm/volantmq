package proxy

import (
	"bytes"
	"errors"
	"github.com/troian/surgemq/message"
	"github.com/troian/surgemq/server"
	"net"
	"sync"
	"time"
)

type client struct {
	id         string
	in         net.Conn
	out        net.Conn
	done       chan struct{}
	wgWorkers  sync.WaitGroup
	signalStop func(id string)
}

type Provider struct {
	ln        net.Listener
	done      chan struct{}
	lock      sync.Mutex
	clients   map[string]*client
	wgClients sync.WaitGroup
}

func NewProxy(from string, to string) (*Provider, error) {
	p := &Provider{
		done:    make(chan struct{}),
		clients: make(map[string]*client),
	}

	var err error
	if p.ln, err = net.Listen("tcp4", from); err != nil {
		return nil, err
	}

	go func() {
		for {
			var in net.Conn
			var err error

			if in, err = p.ln.Accept(); err != nil {
				// http://zhen.org/blog/graceful-shutdown-of-go-net-dot-listeners/
				select {
				case <-p.done:
					return
				default:
				}

				// Borrowed from go1.3.3/src/pkg/net/http/server.go:1699
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					var tempDelay time.Duration // how long to sleep on accept failure

					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if max := 1 * time.Second; tempDelay > max {
						tempDelay = max
					}
					time.Sleep(tempDelay)
					continue
				} else {
					break
				}
			}
			var buf []byte

			if buf, err = server.GetMessageBuffer(in); err != nil {
				continue
			}

			var msg *message.ConnectMessage
			if mP, _, err := message.Decode(buf); err != nil {
				continue
			} else {
				switch m := mP.(type) {
				case *message.ConnectMessage:
					msg = m
				default:
					continue
				}
			}

			var out net.Conn

			if out, err = net.Dial("tcp", to); err != nil {
				in.Close()
				continue
			}
			cl := &client{
				id:         string(msg.ClientID()),
				in:         in,
				out:        out,
				done:       make(chan struct{}),
				signalStop: p.clientStop,
			}

			p.lock.Lock()
			p.clients[string(msg.ClientID())] = cl
			p.wgClients.Add(1)
			p.lock.Unlock()

			cl.serve()

			// Proxy CONNECT message over to MQTT broker
			out.Write(buf)
		}
	}()

	return p, nil
}

func (p *Provider) Shutdown() error {
	select {
	case <-p.done:
		return errors.New("")
	default:
	}

	p.ln.Close()
	return nil
}

func (p *Provider) clientStop(id string) {
	defer p.lock.Unlock()
	p.lock.Lock()

	delete(p.clients, id)
	p.wgClients.Done()
}

func (c *client) serve() {
	c.wgWorkers.Add(2)
	go c.toBroker()
	go c.fromBroker()
}

func (c *client) toBroker() {
	defer func() {
		c.wgWorkers.Done()
		go c.stop()
	}()

	alive := true
	for alive == true {
		var buf []byte
		var err error
		if buf, err = server.GetMessageBuffer(c.in); err != nil {
			continue
		}

		if mP, _, err := message.Decode(buf); err != nil {
			continue
		} else {
			switch m := mP.(type) {
			case *message.PublishMessage:
				if (m.Topic() == "MQTTSAS topic") && (bytes.Compare(m.Payload(), []byte("TERMINATE")) == 0) {
					alive = false
				} else {
					c.out.Write(buf)
				}
			default:
				c.out.Write(buf)
			}
		}
	}
}

func (c *client) fromBroker() {
	defer func() {
		c.wgWorkers.Done()
		go c.stop()
	}()

	buf := make([]byte, 1024)
	for {
		total, err := c.out.Read(buf)
		if total > 0 {
			c.in.Write(buf[:total])
		}

		if err != nil {
			break
		}
	}
}

func (c *client) stop() {
	select {
	case <-c.done:
		return
	default:
		close(c.done)
	}

	c.in.Close()
	c.out.Close()

	c.wgWorkers.Wait()

	c.signalStop(c.id)
}
