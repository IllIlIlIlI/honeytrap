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
package mongodb

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

type MsgHeader struct {
	MessageLength int32
	RequestID     int32
	ResponseTo    int32
	OpCode        int32
}

func (s *mongodbService) decodeMsgHeader(command []byte) MsgHeader {
	mh := MsgHeader{}
	mh.MessageLength = s.decodeInt32(command, 0).(int32)
	mh.RequestID = s.decodeInt32(command, 4).(int32)
	mh.ResponseTo = s.decodeInt32(command, 8).(int32)
	mh.OpCode = s.decodeInt32(command, 12).(int32)
	return mh
}

func (s *mongodbService) decodeInt32(cmd []byte, indexB int16) interface{} {
	a := binary.LittleEndian.Uint32(cmd[indexB:])
	b := fmt.Sprintf("%#04x", a)
	c, _ := strconv.ParseInt(b, 0, 32)
	return int32(c)
}

func (s *mongodbService) decodeOp_msg(cmd []byte) string {
	var rq []byte
	for i := 0; cmd[i] != 0x00; i++ {
		rq = append(rq, cmd[i])
	}
	return string(rq)
}
