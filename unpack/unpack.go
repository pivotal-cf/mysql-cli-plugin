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
	"path/filepath"

	"github.com/gobuffalo/packr"
	"github.com/pkg/errors"
)

//go:generate counterfeiter . Box
type Box interface {
	Walk(walkFunc packr.WalkFunc) error
}

//go:generate go install github.com/pivotal-cf/mysql-cli-plugin/vendor/github.com/gobuffalo/packr/...
//go:generate $GOPATH/bin/packr --compress
var (
	defaultBox = packr.NewBox("../app")
)

type Unpacker struct {
	Box Box
}

func NewUnpacker() *Unpacker {
	return &Unpacker{
		defaultBox,
	}
}

func (u *Unpacker) Unpack(destDir string) error {
	err := u.Box.Walk(func(name string, file packr.File) error {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(destDir, name)), 0700); err != nil {
			return err
		}

		dest, err := os.Create(filepath.Join(destDir, name))
		if err != nil {
			return err
		}

		if _, err := io.Copy(dest, file); err != nil {
			return err
		}

		// todo: works on Windows?
		return dest.Chmod(0700)
	})

	if err != nil {
		return errors.Errorf("Error extracting migrate assets: %s", err)
	}

	return nil
}
