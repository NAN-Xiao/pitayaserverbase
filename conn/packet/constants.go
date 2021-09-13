// Copyright (c) nano Author and TFG Co. All Rights Reserved.
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

package packet

import "errors"

// Type represents the network packet's type such as: handshake and so on.
// 表示网络数据包的类型，如:握手等。
type Type byte

const (
	_ Type = iota
	// Handshake represents a handshake: request(client) <====> handshake response(server)
	// 表示一次握手:request(client) <====>握手响应(server)
	Handshake = 0x01

	// HandshakeAck represents a handshake ack from client to server
	// 表示从客户端到服务器的握手ack
	HandshakeAck = 0x02

	// Heartbeat represents a heartbeat
	// 心跳數據包
	Heartbeat = 0x03

	// Data represents a common data packet
	// 普通的數據包
	Data = 0x04

	// Kick represents a kick off packet
	// 代表开球包
	Kick = 0x05 // disconnect message from server
)

// ErrWrongPomeloPacketType represents a wrong packet type.
// 錯誤的包類型
var ErrWrongPomeloPacketType = errors.New("wrong packet type")

// ErrInvalidPomeloHeader represents an invalid header
// 無效的頭
var ErrInvalidPomeloHeader = errors.New("invalid header")
