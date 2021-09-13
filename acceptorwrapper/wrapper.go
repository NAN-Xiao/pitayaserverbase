// Copyright (c) TFG Co. All Rights Reserved.
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

package acceptorwrapper

import (
	"github.com/topfreegames/pitaya/v2/acceptor"
)

// Wrapper has a method that receives an acceptor and the struct
// that implements must encapsulate it. The main goal is to create
// a middleware for packets of net.Conn from acceptor.GetConnChan before
// giving it to serviceHandler.
// 包装器有一个方法接收一个接受器和结构体
// 实现的对象必须封装它。主要目标是创造
// 网络包的中间件。GetConnChan之前
// 将它赋给serviceHandler
type Wrapper interface {
	Wrap(acceptor.Acceptor) acceptor.Acceptor
}

// WithWrappers walks through wrappers calling Wrapper
func WithWrappers(
	a acceptor.Acceptor,
	wrappers ...Wrapper,
) acceptor.Acceptor {
	for _, w := range wrappers {
		a = w.Wrap(a)
	}
	return a
}
