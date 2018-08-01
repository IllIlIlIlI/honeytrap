/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */

package couchdb

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("services/couchdb")
	_   = services.Register("couchdb", Couchdb)
)

func Couchdb(options ...services.ServicerFunc) services.Servicer {
	s := &couchdbService{}

	for _, o := range options {
		o(s)
	}
	return s
}

type couchdbService struct {
	ch pushers.Channel
}

func (s *couchdbService) SetChannel(c pushers.Channel) {
	s.ch = c
}

func Headers(headers map[string][]string) event.Option {
	return func(m event.Event) {
		for name, h := range headers {
			m.Store(fmt.Sprintf("http.header.%s", strings.ToLower(name)), h)
		}
	}
}

var couchdbMethods = map[string]func() interface{}{
	"/_all_dbs": func() interface{} {
		return []interface{}{
			"_replicator",
			"_users",
		}
	},
}

func (s *couchdbService) Handle(ctx context.Context, conn net.Conn) error {

	defer conn.Close()

	br := bufio.NewReader(conn)

	req, err := http.ReadRequest(br)
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(req.Body)
	if err == io.EOF {
	} else if err != nil {
		return err
	}

	s.ch.Send(event.New(
		services.EventOptions,
		event.Category("couchdb"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("http.user-agent", req.UserAgent()),
		event.Custom("http.method", req.Method),
		event.Custom("http.proto", req.Proto),
		event.Custom("http.host", req.Host),
		event.Custom("http.url", req.URL.String()),
		event.Custom("couchdb.method", req.URL.Path),
		event.Payload(data),
		Headers(req.Header),
	))

	buff := bytes.Buffer{}

	fn, ok := couchdbMethods[req.RequestURI]
	if ok {
		v := fn()

		if err := json.NewEncoder(&buff).Encode(v); err != nil {
			return err
		}
	} else {
		log.Errorf("Method %s not supported", req.URL.Path)
	}

	resp := http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Request:    req,
		Header:     http.Header{},
	}

	resp.Header.Add("Content-Length", fmt.Sprintf("%d", buff.Len()))
	resp.Header.Add("Content-Type", "text/plain; charset=UTF-8")
	resp.Header.Add("Cache-Control", "must-revalidate")
	resp.Header.Add("Server", "CouchDB/1.7.1 (Erlang OTP/19)")
	resp.Header.Add("Date", time.Now().Format("Mon, _2 Jan 2006 15:04:05 GMT"))

	resp.Body = ioutil.NopCloser(&buff)

	return resp.Write(conn)

}
