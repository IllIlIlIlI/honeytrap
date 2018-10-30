/*
* Honeytrap
* Copyright (C) 2016-2018 DutchSec (https://dutchsec.com/)
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

/*
[service.couchdb]
type=“couchdb”

[[port]]
port="tcp/5984”
services=[“couchdb”]
*/

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

func (s *couchdbService) CanHandle(payload []byte) bool {
	if bytes.HasPrefix(payload, []byte("GET")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("HEAD")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("POST")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("PUT")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("DELETE")) {
		return true
	}
	return false
}

func Headers(headers map[string][]string) event.Option {
	return func(m event.Event) {
		for name, h := range headers {
			m.Store(fmt.Sprintf("http.header.%s", strings.ToLower(name)), h)
		}
	}
}

var couchdbMethods = map[string]func() interface{}{
	"/": func() interface{} {
		return map[string]interface{}{
			"couchdb": "Welcome",
			"uuid":    "9cee8576d97567a5e155cf78ff041785",
			"version": "1.7.1",
			"vendor": map[string]string{
				"name":    "Homebrew",
				"version": "1.7.1_7",
			},
		}
	},
	"/_all_dbs": func() interface{} {
		return []string{
			"_replicator",
			"_users",
		}
	},
	"/_config": func() interface{} {
		return map[string]interface{}{
			"httpd_design_handlers": map[string]string{
				"_compact": "{couch_mrview_http, handle_compact_req}",
				"_info":    "{couch_mrview_http, handle_info_req}",
				"_list":    "{couch_mrview_show, handle_view_list_req}",
				"_rewrite": "{couch_httpd_rewrite, handle_rewrite_req}",
				"_show":    "{couch_mrview_show, handle_doc_show_req}",
				"_update":  "{couch_mrview_show, handle_doc_update_req}",
				"_view":    "{couch_mrview_http, handle_view_req}",
			},
			"uuids": map[string]string{
				"algorithm": "sequential",
				"max_count": "1000",
			},
			"stats": map[string]string{
				"rate":    "1000",
				"samples": "[0, 60, 300, 900]",
			},
			"cors": map[string]string{
				"credentials": "false",
			},
			"httpd_global_handlers": map[string]string{
				"/":             "{couch_httpd_misc_handlers, handle_welcome_req, <<\"Welcome\">>}",
				"_active_tasks": "{couch_httpd_misc_handlers, handle_task_status_req}",
				"_all_dbs":      "{couch_httpd_misc_handlers, handle_all_dbs_req}",
				"_config":       "{couch_httpd_misc_handlers, handle_config_req}",
				"_db_updates":   "{couch_dbupdates_httpd, handle_req}",
				"_log":          "{couch_httpd_misc_handlers, handle_log_req}",
				"_oauth":        "{couch_httpd_oauth, handle_oauth_req}",
				"_plugins":      "{couch_plugins_httpd, handle_req}",
				"_replicate":    "{couch_replicator_httpd, handle_req}",
				"_restart":      "{couch_httpd_misc_handlers, handle_restart_req}",
				"_session":      "{couch_httpd_auth, handle_session_req}",
				"_stats":        "{couch_httpd_stats_handlers, handle_stats_req}",
				"_utils":        "{couch_httpd_misc_handlers, handle_utils_dir_req, \"/usr/local/Cellar/couchdb/1.7.1_7/share/couchdb/www\"}",
				"_uuids":        "{couch_httpd_misc_handlers, handle_uuids_req}",
				"favicon.ico":   "{couch_httpd_misc_handlers, handle_favicon_req, \"/usr/local/Cellar/couchdb/1.7.1_7/share/couchdb/www\"}",
			},
			"attachments": map[string]string{
				"compressible_types": "text/*, application/javascript, application/json, application/xml",
				"compression_level":  "8",
			},
			"query_server_config": map[string]string{
				"os_process_limit": "25",
				"reduce_limit":     "true",
			},
			"vendor": map[string]string{
				"name":    "Homebrew",
				"version": "1.7.1_7",
			},
			"replicator": map[string]string{
				"connection_timeout":          "30000",
				"db":                          "_replicator",
				"http_connections":            "20",
				"max_replication_retry_count": "10",
				"retries_per_request":         "10",
				"socket_options":              "[{keepalive, true}, {nodelay, false}]",
				"ssl_certificate_max_depth":   "3",
				"verify_ssl_certificates":     "false",
				"worker_batch_size":           "500",
				"worker_processes":            "4",
			},
			"couch_httpd_oauth": map[string]string{
				"use_users_db": "false",
			},
			"ssl": map[string]string{
				"port": "6984",
				"ssl_certificate_max_depth": "1",
				"verify_ssl_certificates":   "false",
			},
			"log": map[string]string{
				"file":         "/usr/local/var/log/couchdb/couch.log",
				"include_sasl": "true",
				"level":        "info",
			},
			"view_compaction": map[string]string{
				"keyvalue_buffer_size": "2097152",
			},
			"query_servers": map[string]string{
				"coffeescript": "/usr/local/Cellar/couchdb/1.7.1_7/bin/couchjs /usr/local/Cellar/couchdb/1.7.1_7/share/couchdb/server/main-coffee.js",
				"javascript":   "/usr/local/Cellar/couchdb/1.7.1_7/bin/couchjs /usr/local/Cellar/couchdb/1.7.1_7/share/couchdb/server/main.js",
			},
			"daemons": map[string]string{
				"auth_cache":         "{couch_auth_cache, start_link, []}",
				"compaction_daemon":  "{couch_compaction_daemon, start_link, []}",
				"external_manager":   "{couch_external_manager, start_link, []}",
				"httpd":              "{couch_httpd, start_link, []}",
				"index_server":       "{couch_index_server, start_link, []}",
				"os_daemons":         "{couch_os_daemons, start_link, []}",
				"query_servers":      "{couch_query_servers, start_link, []}",
				"replicator_manager": "{couch_replicator_manager, start_link, []}",
				"stats_aggregator":   "{couch_stats_aggregator, start, []}",
				"stats_collector":    "{couch_stats_collector, start, []}",
				"uuids":              "{couch_uuids, start, []}",
				"vhosts":             "{couch_httpd_vhost, start_link, []}",
			},
			"httpd": map[string]string{
				"allow_jsonp":             "false",
				"authentication_handlers": "{couch_httpd_oauth, oauth_authentication_handler}, {couch_httpd_auth, cookie_authentication_handler}, {couch_httpd_auth, default_authentication_handler}",
				"bind_address":            "0.0.0.0",
				"default_handler":         "{couch_httpd_db, handle_request}",
				"enable_cors":             "false",
				"log_max_chunk_size":      "1000000",
				"port":                    "5984",
				"secure_rewrites":         "true",
				"socket_options":          "[{recbuf, 262144}, {sndbuf, 262144}, {nodelay, true}]",
				"vhost_global_handlers":   "_utils, _uuids, _session, _oauth, _users",
			},
			"httpd_db_handlers": map[string]string{
				"_all_docs":     "{couch_mrview_http, handle_all_docs_req}",
				"_changes":      "{couch_httpd_db, handle_changes_req}",
				"_compact":      "{couch_httpd_db, handle_compact_req}",
				"_design":       "{couch_httpd_db, handle_design_req}",
				"_temp_view":    "{couch_mrview_http, handle_temp_view_req}",
				"_view_cleanup": "{couch_mrview_http, handle_cleanup_req}",
			},
			"database_compaction": map[string]string{
				"checkpoint_after": "5242880",
				"doc_buffer_size":  "524288",
			},
			"couch_httpd_auth": map[string]string{
				"allow_persistent_cookies": "false",
				"auth_cache_size":          "50",
				"authentication_db":        "_users",
				"authentication_redirect":  "/_utils/session.html",
				"iterations":               "10",
				"require_valid_user":       "false",
				"timeout":                  "600",
			},
			"couchdb": map[string]string{
				"attachment_stream_buffer_size": "4096",
				"database_dir":                  "/usr/local/var/lib/couchdb",
				"delayed_commits":               "true",
				"file_compression":              "snappy",
				"max_dbs_open":                  "100",
				"max_document_size":             "4294967296",
				"os_process_timeout":            "5000",
				"plugin_dir":                    "/usr/local/Cellar/couchdb/1.7.1_7/lib/couchdb/plugins",
				"uri_file":                      "/usr/local/var/run/couchdb/couch.uri",
				"util_driver_dir":               "/usr/local/Cellar/couchdb/1.7.1_7/lib/couchdb/erlang/lib/couch-1.7.1/priv/lib",
				"uuid":                          "9cee8576d97567a5e155cf78ff041785",
				"view_index_dir":                "/usr/local/var/lib/couchdb",
			},
			"compaction_daemon": map[string]string{
				"check_interval": "300",
				"min_file_size":  "131072",
			},
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
	resp.Header.Add("Date", time.Now().Format("Mon, 2 Jan 2006 15:04:05 GMT"))

	resp.Body = ioutil.NopCloser(&buff)

	return resp.Write(conn)

}
