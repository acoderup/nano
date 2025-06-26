// Copyright (c) nano Authors. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/acoderup/nano/internal/packet"
)

// Codec constants.
const (
	HeadLength    = 4
	MaxPacketSize = 64 * 1024
)

// ErrPacketSizeExcced is the error used for encode/decode.
var ErrPacketSizeExcced = errors.New("codec: packet size exceed")

// A Decoder reads and decodes network data slice
type Decoder struct {
	buf  *bytes.Buffer
	size int  // last packet length
	typ  byte // last packet type
}

// NewDecoder returns a new decoder that used for decode network bytes slice.
func NewDecoder() *Decoder {
	return &Decoder{
		buf:  bytes.NewBuffer(nil),
		size: -1,
	}
}

// Decode decode the network bytes slice to packet.Packet(s)
// TODO(Warning): shared slice
func (c *Decoder) Decode(buffer []byte) ([]*packet.Packet, error) {
	// 将新数据写入缓冲区
	if len(buffer) > 0 {
		c.buf.Write(buffer)
	}

	var packets []*packet.Packet

	for {
		// 如果需要解析新包头 (c.size == -1 表示需要读取新包头)
		if c.size == -1 {
			// 检查是否有足够数据读取头部基本信息
			if c.buf.Len() < 1 {
				break
			}

			// 读取第一个字节但不移动指针(Peek操作)
			ne := c.buf.Bytes()[0]
			isUint32Len := (ne & 0x08) != 0
			headerLength := 3 // 默认UInt16长度
			if isUint32Len {
				headerLength = 5 // UInt32长度
			}

			// 检查是否有完整头部
			if c.buf.Len() < headerLength {
				break
			}

			// 解析包长度
			if isUint32Len {
				c.size = headerLength + int(binary.BigEndian.Uint32(c.buf.Bytes()[1:5]))
			} else {
				c.size = headerLength + int(binary.BigEndian.Uint16(c.buf.Bytes()[1:3]))
			}

			//// 检查包长度限制
			//if c.size > MaxPacketSize {
			//	// 丢弃错误数据包
			//	c.buf.Next(headerLength)
			//	c.size = -1
			//	return packets, ErrPacketSizeExceed
			//}

			// 移动指针跳过头部
			//c.buf.Next(headerLength)
		}

		// 检查是否有足够数据读取完整包体
		if c.buf.Len() < c.size {
			break
		}

		// 读取包体数据
		packetData := c.buf.Next(c.size)

		// 创建数据包
		packets = append(packets, &packet.Packet{
			Type:   packet.Type(4), // 根据实际协议调整
			Length: c.size,
			Data:   packetData,
		})

		// 重置状态，准备解析下一个包
		c.size = -1
	}

	return packets, nil
}

// Encode create a packet.Packet from  the raw bytes slice and then encode to network bytes slice
// Protocol refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
//
// -<type>-|--------<length>--------|-<data>-
// --------|------------------------|--------
// 1 byte packet type, 3 bytes packet data length(big end), and data segment
func Encode(typ packet.Type, data []byte) ([]byte, error) {
	if typ < packet.Handshake || typ > packet.Kick {
		return nil, packet.ErrWrongPacketType
	}

	p := &packet.Packet{Type: typ, Length: len(data)}
	buf := make([]byte, p.Length+HeadLength)
	buf[0] = byte(p.Type)

	copy(buf[1:HeadLength], intToBytes(p.Length))
	copy(buf[HeadLength:], data)

	return buf, nil
}

// Decode packet data length byte to int(Big end)
func bytesToInt(b []byte) int {
	result := 0
	for _, v := range b {
		result = result<<8 + int(v)
	}
	return result
}

// Encode packet data length to bytes(Big end)
func intToBytes(n int) []byte {
	buf := make([]byte, 3)
	buf[0] = byte((n >> 16) & 0xFF)
	buf[1] = byte((n >> 8) & 0xFF)
	buf[2] = byte(n & 0xFF)
	return buf
}
