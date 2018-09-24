package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"
)

type Extension func(message *Message)

type Extensions struct {
	In  []Extension
	Out []Extension
}

func (e *Extensions) ApplyOutExtensions(m *Message) {
	for i := range e.Out {
		e.Out[i](m)
	}
}

func (e *Extensions) ApplyInExtensions(m *Message) {
	for i := range e.In {
		e.In[i](m)
	}
}

type Data = interface{}

type Message struct {
	Channel                  string      `json:"channel,omitempty"`
	Version                  string      `json:"version,omitempty"`
	SupportedConnectionTypes []string    `json:"supportedConnectionTypes,omitempty"`
	ConnectionType           string      `json:"connectionType,omitempty"`
	MinimumVersion           string      `json:"minimumVersion,omitempty"`
	Successful               bool        `json:"successful,omitempty"`
	Ext                      interface{} `json:"ext,omitempty"`
	Id                       string      `json:"id,omitempty"`
	ClientId                 string      `json:"clientId,omitempty"`
	Advice                   *Advise     `json:"advice,omitempty"`
	Data                     Data        `json:"data,omitempty"`
	Timestamp                uint64      `json:"timestamp,omitempty"`
	AuthSuccessful           bool        `json:"authSuccessful,omitempty"`
	Error                    string      `json:"error,omitempty"`
	Subscription             string      `json:"subscription,omitempty"`
}

func (m *Message) GetError() error {
	if m.Error == "" {
		return nil
	}
	return errors.New(m.Error)
}

type Reconnect string

const (
	//ReconnectRetry indicates that a client MAY attempt to reconnect with a /meta/connect message,
	//after the interval (as defined by interval advice field or client-default backoff), and with the same credentials.
	ReconnectRetry Reconnect = "retry"

	//ReconnectHandshake indicates that the server has terminated any prior connection status and the client MUST reconnect
	// with a /meta/handshake message.
	//A client MUST NOT automatically retry when a reconnect advice handshake has been received.
	ReconnectHandshake Reconnect = "handshake"

	//ReconnectNone indicates a hard failure for the connect attempt.
	//A client MUST respect reconnect advice none and MUST NOT automatically retry or handshake.
	ReconnectNone Reconnect = "none"
)

type Advise struct {
	Reconnect       Reconnect     `json:"reconnect,omitempty"`
	Interval        time.Duration `json:"interval,omitempty"`
	Timeout         time.Duration `json:"timeout,omitempty"`
	MultipleClients bool          `json:"multiple-clients,omitempty"`
	Hosts           []string      `json:"hosts,omitempty"`
}

func (a *Advise) MarshalJSON() ([]byte, error) {
	type jsonStruct struct {
		Reconnect       string   `json:"reconnect,omitempty"`
		Interval        int64    `json:"interval,omitempty"`
		Timeout         int64    `json:"timeout,omitempty"`
		MultipleClients bool     `json:"multiple-clients,omitempty"`
		Hosts           []string `json:"hosts,omitempty"`
	}

	var builder bytes.Buffer

	err := json.NewEncoder(&builder).Encode(jsonStruct{
		Reconnect:       string(a.Reconnect),
		Interval:        int64(a.Interval / time.Millisecond),
		Timeout:         int64(a.Timeout / time.Millisecond),
		MultipleClients: a.MultipleClients,
		Hosts:           a.Hosts,
	})

	return builder.Bytes(), err
}

func (a *Advise) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	reconnect, ok := raw["reconnect"]
	if ok {
		a.Reconnect = Reconnect(reconnect.(string))
	}

	interval, ok := raw["interval"]
	if ok {
		a.Interval = time.Duration(interval.(float64)) * time.Millisecond
	}

	timeout, ok := raw["timeout"]
	if ok {
		a.Timeout = time.Duration(timeout.(float64)) * time.Millisecond
	}

	mc, ok := raw["multiple-clients"]
	if ok {
		a.MultipleClients = mc.(bool)
	}

	hosts, ok := raw["hosts"]
	if ok {
		a.Hosts = hosts.([]string)
	}

	return nil
}
func IsEventDelivery(msg *Message) bool {
	if IsMetaMessage(msg) {
		return false
	}
	if msg.Data != nil {
		return true
	}
	return false
}

//MetaMessage are channels commencing with the /meta/ segment ans, are the channels used by the faye protocol itself.
type MetaMessage = string

const (
	MetaSubscribe   MetaMessage = "/meta/subscribe"
	MetaConnect     MetaMessage = "/meta/connect"
	MetaDisconnect  MetaMessage = "/meta/disconnect"
	MetaUnsubscribe MetaMessage = "/meta/unsubscribe"
	MetaHandshake   MetaMessage = "/meta/handshake"
)

//EventMessage are published in event messages sent from a faye client to a faye server
//and are delivered in event messages sent from a faye server to a faye client.
type EventMessage = int

const (
	//
	EventPublish EventMessage = iota
	EventDelivery
)

var metaMessages = []MetaMessage{MetaSubscribe, MetaConnect, MetaUnsubscribe, MetaHandshake, MetaDisconnect}

func IsMetaMessage(msg *Message) bool {
	for i := range metaMessages {
		if msg.Channel == metaMessages[i] {
			return true
		}
	}
	return false
}

func IsEventPublish(msg *Message) bool {
	if IsMetaMessage(msg) {
		return false
	}
	return !IsEventDelivery(msg)
}
