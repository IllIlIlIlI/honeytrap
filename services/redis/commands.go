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
package redis

import (
	"fmt"
)

type cmd func(*redisService, []interface{}) (string, bool)

var mapCmds = map[string]cmd{
	"info":     (*redisService).infoCmd,
	"flushall": (*redisService).flushallCmd,
	"save":     (*redisService).saveCmd,
	"set":      (*redisService).setCmd,
	"config":   (*redisService).configCmd,
	// ...
}

type infoSection func(*redisService) string

var mapInfoCmds = map[string]infoSection{
	"server":      (*redisService).infoServerMsg,
	"clients":     (*redisService).infoClientsMsg,
	"memory":      (*redisService).infoMemoryMsg,
	"persistence": (*redisService).infoPersistenceMsg,
	"stats":       (*redisService).infoStatsMsg,
	"replication": (*redisService).infoReplicationMsg,
	"cpu":         (*redisService).infoCPUMsg,
	"cluster":     (*redisService).infoClusterMsg,
	"keyspace":    (*redisService).infoKeyspaceMsg,
}

var mapConfigCmds = map[string]cmd{
	"get":       (*redisService).configGetCmd,
	"set":       (*redisService).configSetCmd,
	"resetstat": (*redisService).configResetstatCmd,
	"rewrite":   (*redisService).configRewriteCmd,
}

func (s *redisService) infoCmd(args []interface{}) (string, bool) {
	switch len(args) {
	case 0:
		return fmt.Sprintf(lenMsg(), len(s.infoSectionsMsg()), s.infoSectionsMsg()), false
	case 1:
		_word := args[0].(redisDatum)
		word, success := _word.ToString()
		if !success {
			return "Expected string argument, got something else", false
		}
		if fn, ok := mapInfoCmds[word]; ok {
			return fmt.Sprintf(lenMsg(), len(fn(s)), fn(s)), false
		} else if word == "default" {
			return fmt.Sprintf(lenMsg(), len(s.infoSectionsMsg()), s.infoSectionsMsg()), false
		} else if word == "all" {
			return fmt.Sprintf(lenMsg(), len(s.allSectionsMsg()), s.allSectionsMsg()), false
		} else {
			return fmt.Sprintf(lenMsg(), len(lineBreakMsg()), lineBreakMsg()), false
		}
	default:
		return errorMsg("syntax"), false
	}
}

func (s *redisService) flushallCmd(args []interface{}) (string, bool) {
	switch len(args) {
	case 0:
		return "+OK\r\n", false
	case 1:
		_word := args[0].(redisDatum)
		word, success := _word.ToString()
		if !success {
			return "Expected string argument, got something else", false
		}
		if word == "async" {
			return "+OK\r\n", false
		}
		fallthrough
	default:
		return errorMsg("syntax"), false

	}
}

func (s *redisService) saveCmd(args []interface{}) (string, bool) {
	switch len(args) {
	case 0:
		return "+OK\r\n", false
	default:
		return fmt.Sprintf(errorMsg("wgnumber"), "save"), false
	}
}

func (s *redisService) setCmd(args []interface{}) (string, bool) {
	switch len(args) {
	case 2:
		return "+OK\r\n", false
	case 0:
		fallthrough
	case 1:
		return fmt.Sprintf(errorMsg("wgnumber"), "set"), false
	default:
		return errorMsg("syntax"), false
	}
}

func (s *redisService) configCmd(args []interface{}) (string, bool) {
	if len(args) == 0 {
		return fmt.Sprintf(errorMsg("wgnumber"), "config"), false
	}

	_word := args[0].(redisDatum)
	word, success := _word.ToString()
	if !success {
		return "Expected string argument, got something else", false
	}

	switch word {
	case "get":
		return s.configGetCmd(args)
	case "set":
		return s.configSetCmd(args)
	case "resetstat":
		return s.configResetstatCmd(args)
	case "rewrite":
		return s.configRewriteCmd(args)
	default:
		return errorConfig("config"), false
	}
}

func (s *redisService) configGetCmd(args []interface{}) (string, bool) {
	switch len(args) {
	case 2:
		// [...]
		return "*0\r\n", false
	default:
		return fmt.Sprintf(errorConfig("wgnumber"), "get"), false
	}
}

func (s *redisService) configSetCmd(args []interface{}) (string, bool) {
	switch len(args) {

	case 3:
		// [check parameters]
		return "+OK\r\n", false
	default:
		return fmt.Sprintf(errorConfig("wgnumber"), "set"), false
	}
}

func (s *redisService) configResetstatCmd(args []interface{}) (string, bool) {
	switch len(args) {
	case 0:
		// [linked to INFO]
		return "+OK\r\n", false
	default:
		return fmt.Sprintf(errorConfig("wgnumber"), "resetstat"), false
	}
}

func (s *redisService) configRewriteCmd(args []interface{}) (string, bool) {
	switch len(args) {
	case 0:
		return "+OK\r\n", false
	default:
		return fmt.Sprintf(errorConfig("wgnumber"), "rewrite"), false
	}
}
