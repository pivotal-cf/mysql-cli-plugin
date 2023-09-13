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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"

	"github.com/gobuffalo/packr"
)

//go:generate packr --compress
var (
	defaultBox = packr.NewBox("../../app")
)

//counterfeiter:generate . Box
type Box interface {
	Walk(walkFunc packr.WalkFunc) error
}

type Unpacker struct {
	Box        Box
	Filesystem Filesystem
}

func NewUnpacker() *Unpacker {
	return &Unpacker{
		Box:        defaultBox,
		Filesystem: LocalFilesystem{},
	}
}

func (u *Unpacker) Unpack(destDir string) error {
	err := u.Box.Walk(func(name string, file packr.File) error {
		if err := u.Filesystem.MkdirAll(filepath.Dir(filepath.Join(destDir, name)), 0o700); err != nil {
			return err
		}

		dest, err := u.Filesystem.Create(filepath.Join(destDir, name))
		if err != nil {
			return err
		}
		defer dest.Close()

		if _, err := io.Copy(dest, file); err != nil {
			return err
		}

		if runtime.GOOS != "windows" {
			if err := dest.Chmod(0o700); err != nil {
				return err
			}
		}

		return dest.Close()
	})
	if err != nil {
		return fmt.Errorf("Error extracting migrate assets: %s", err)
	}

	return nil
}
