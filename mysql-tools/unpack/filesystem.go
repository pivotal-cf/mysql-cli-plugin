// Copyright (C) 2018-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under the terms of the under the Apache License,
// Version 2.0 (the "License”); you may not use this file except in compliance with the License. You may obtain a copy
// of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package unpack

import (
	"io"
	"os"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . File
type File interface {
	io.ReadWriteCloser
	Chmod(mode os.FileMode) error
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Filesystem
type Filesystem interface {
	Create(name string) (File, error)
	MkdirAll(path string, perm os.FileMode) error
}

type LocalFilesystem struct{}

func (LocalFilesystem) Create(name string) (File, error) {
	return os.Create(name)
}

func (LocalFilesystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
