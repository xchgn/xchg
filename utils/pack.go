// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Xchg-Network Authors
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

package utils

import (
	"archive/zip"
	"bytes"
	"io"
	"io/fs"
)

func PackBytes(data []byte) []byte {
	var err error
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	var zipFile io.Writer
	zipFile, err = zipWriter.Create("data")
	if err == nil {
		_, err = zipFile.Write(data)
		if err != nil {
			return nil
		}
	}
	err = zipWriter.Close()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

func UnpackBytes(zippedData []byte) (result []byte, err error) {
	buf := bytes.NewReader(zippedData)
	var zipFile *zip.Reader
	zipFile, err = zip.NewReader(buf, buf.Size())
	if err != nil {
		return
	}
	var file fs.File
	file, err = zipFile.Open("data")
	if err == nil {
		result, err = io.ReadAll(file)
		_ = file.Close()
	}
	return
}
