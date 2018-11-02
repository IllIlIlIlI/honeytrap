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
package claymore

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

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	logging "github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("services/claymore")
	_ = services.Register("claymore", Claymore)
)


/*

[service.claymore]
type="claymore"

[[port]]
port="tcp/3333"
services=["claymore"]

*/

func Claymore(options ...services.ServicerFunc) services.Servicer {
	s := &claymoreService{}

	for _, o := range options {
		o(s)
	}

	return s
}

type claymoreService struct {
	ch pushers.Channel
}

func (s *claymoreService) SetChannel(c pushers.Channel) {
	s.ch = c
}

func Headers(headers map[string][]string) event.Option {
	return func(m event.Event) {
		for name, h := range headers {
			m.Store(fmt.Sprintf("http.header.%s", strings.ToLower(name)), h)
		}
	}
}

var claymoreMethods = map[string]func(map[string]interface{}) map[string]interface{}{

	"miner_reboot": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result":  true,
		}
	},
	"miner_file": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result":  true,
		}
	},
	"miner_getstat1": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result": []string{
				"10.1 - ETH",
				"4286",
				"149336;7492;0",
				"30620;29877;28285;30605;29946",
				"0;0;0",
				"off;off;off;off;off",
				"62;65;51;64;61;75;51;67;62;72",
				"eth-eu1.nanopool.org:9999",
				"0;1;0;0",
			},
		}
	},
}

func (s *claymoreService) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	br := bufio.NewReader(conn)

	req, err := http.ReadRequest(br)
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	jsonRequest := map[string]interface{}{}
	if err := json.Unmarshal(data, &jsonRequest); err != nil {
		return err
	}

	method := ""

	if s, ok := jsonRequest["method"].(string); ok {
		method = s
	}

	s.ch.Send(event.New(
		services.EventOptions,
		event.Category("claymore"),
		event.Type(method),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("http.user-agent", req.UserAgent()),
		event.Custom("http.method", req.Method),
		event.Custom("http.proto", req.Proto),
		event.Custom("http.host", req.Host),
		event.Custom("http.url", req.URL.String()),
		event.Custom("claymore.id", jsonRequest["id"]),
		event.Custom("claymore.method", jsonRequest["method"]),
		event.Custom("claymore.jsonrpc", jsonRequest["jsonrpc"]),
		event.Payload(data),
		Headers(req.Header),
	))

	buff := bytes.Buffer{}

	fn, ok := claymoreMethods[method]
	if ok {
		v := fn(jsonRequest)

		if err := json.NewEncoder(&buff).Encode(v); err != nil {
			return err
		}
	} else {
		log.Errorf("Method %s not supported", method)
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

	resp.Header.Add("Content-Type", "application/json; charset=UTF-8")
	resp.Header.Add("Content-Length", fmt.Sprintf("%d", buff.Len()))

	resp.Body = ioutil.NopCloser(&buff)

	return resp.Write(conn)
}
