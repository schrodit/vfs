/*
 * Copyright 2020 Mandelsoft. All rights reserved.
 *  This file is licensed under the Apache Software License, v. 2 except as noted
 *  otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package memory

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/mandelsoft/vfs/pkg/test"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

var _ = Describe("memory filesystem", func() {
	var fs vfs.FileSystem

	BeforeEach(func() {
		fs = New()
	})

	Context("dirs", func() {

		It("empty", func() {
			ExpectFolders(fs, "/", nil, nil)
		})

		It("create dir", func() {
			Expect(fs.Mkdir("d1", os.ModePerm)).To(BeNil())
			ExpectFolders(fs, "/", []string{"d1"}, nil)
			ExpectFolders(fs, "/d1", nil, nil)
		})
		It("create 2 dirs", func() {
			Expect(fs.Mkdir("d2", os.ModePerm)).To(BeNil())
			Expect(fs.Mkdir("d1", os.ModePerm)).To(BeNil())
			ExpectFolders(fs, "/", []string{"d1", "d2"}, nil)
			ExpectFolders(fs, "/d1", nil, nil)
			ExpectFolders(fs, "/d2", nil, nil)
		})
		It("remove root", func() {
			Expect(fs.Remove("/")).To(Equal(errors.New("cannot delete root dir")))
		})
		It("remove nested", func() {
			Expect(fs.Mkdir("d1", os.ModePerm)).To(BeNil())
			Expect(fs.Mkdir("d1/d2", os.ModePerm)).To(BeNil())
			ExpectFolders(fs, "/d1", []string{"d2"}, nil)
			Expect(fs.Remove("/d1")).To(Equal(&os.PathError{"remove", "/d1", ErrNotEmpty}))
			Expect(fs.Remove("/d1/d2")).To(Succeed())
			ExpectFolders(fs, "/d1", []string{}, nil)
		})
		It("mkdirall", func() {
			Expect(fs.MkdirAll("d1/d2/d3", os.ModePerm)).To(Succeed())
			ExpectFolders(fs, "/", []string{"d1"}, nil)
			ExpectFolders(fs, "/d1", []string{"d2"}, nil)
			ExpectFolders(fs, "/d1/d2", []string{"d3"}, nil)
			ExpectFolders(fs, "/d1/d2/d3", nil, nil)
		})
		It("partial mkdirall", func() {
			Expect(fs.MkdirAll("d1/d2", os.ModePerm)).To(Succeed())
			Expect(fs.MkdirAll("d1/d2/d3/d4", os.ModePerm)).To(Succeed())
			ExpectFolders(fs, "/", []string{"d1"}, nil)
			ExpectFolders(fs, "/d1", []string{"d2"}, nil)
			ExpectFolders(fs, "/d1/d2", []string{"d3"}, nil)
			ExpectFolders(fs, "/d1/d2/d3", []string{"d4"}, nil)
			ExpectFolders(fs, "/d1/d2/d3/d4", nil, nil)
		})

	})

	Context("dots", func() {
		It("dot", func() {
			Expect(fs.MkdirAll("d1/d2", os.ModePerm)).To(BeNil())
			ExpectFolders(fs, ".", []string{"d1"}, nil)
			ExpectFolders(fs, "d1/.", []string{"d2"}, nil)
		})
		It("dotdot", func() {
			Expect(fs.Mkdir("d1", os.ModePerm)).To(BeNil())
			ExpectFolders(fs, "..", []string{"d1"}, nil)
			ExpectFolders(fs, "d1/..", []string{"d1"}, nil)
			ExpectFolders(fs, "d1/../..", []string{"d1"}, nil)
			ExpectFolders(fs, "../d1/../..", []string{"d1"}, nil)
		})
	})

	Context("files", func() {
		It("create file in dir", func() {
			Expect(fs.Mkdir("d1", os.ModePerm)).To(BeNil())
			ExpectCreateFile(fs, "/d1/f1", nil, nil)
		})
		It("read", func() {
			ExpectCreateFile(fs, "f1", []byte("This is a test\n"), nil)
			f, err := fs.Open("/f1")
			Expect(err).To(Succeed())
			ExpectRead(f, []byte("This is a test\n"))
		})
	})
	Context("symlinks", func() {
		BeforeEach(func() {
			fs.MkdirAll("d1/d1n1/d1n1a", os.ModePerm)
			fs.MkdirAll("d1/d1n2", os.ModePerm)
			fs.MkdirAll("d2/d2n1", os.ModePerm)
			fs.MkdirAll("d2/d2n2", os.ModePerm)
		})

		It("creates link", func() {
			Expect(fs.Symlink("/d1/d1n1", "d2/link")).To(Succeed())
			Expect(fs.Readlink("/d2/link")).To(Equal("/d1/d1n1"))
		})

		It("lstat link", func() {
			Expect(fs.Symlink("/d1/d1n1", "d2/link")).To(Succeed())
			fi, err := fs.Lstat("d2/link")
			Expect(err).To(Succeed())
			Expect(fi.Mode() & os.ModeType).To(Equal(os.ModeSymlink))
		})

		It("stat link", func() {
			Expect(fs.Symlink("/d1/d1n1", "d2/link")).To(Succeed())
			fi, err := fs.Stat("d2/link")
			Expect(err).To(Succeed())
			Expect(fi.Mode() & os.ModeType).To(Equal(os.ModeDir))
		})

		Context("eval", func() {
			It("plain", func() {
				Expect(fs.Symlink("/d1/d1n1", "d2/link")).To(Succeed())
				ExpectFolders(fs, "d2/link", []string{"d1n1a"}, nil)
			})
			It("dotdot", func() {
				Expect(fs.Symlink("/d1/d1n1", "d2/link")).To(Succeed())
				ExpectFolders(fs, "d2/link/..", []string{"d1n1", "d1n2"}, nil)
			})
			It("dotdot in link", func() {
				Expect(fs.Symlink("../d1", "d2/link")).To(Succeed())
				ExpectFolders(fs, "d2/link", []string{"d1n1", "d1n2"}, nil)
			})
		})
	})
	Context("rename", func() {
		BeforeEach(func() {
			fs.MkdirAll("d1/d1n1/d1n1a", os.ModePerm)
			fs.MkdirAll("d1/d1n2", os.ModePerm)
		})
		It("rename top level", func() {
			Expect(fs.Rename("/d1", "d2")).To(Succeed())
			ExpectFolders(fs, "d2/", []string{"d1n1", "d1n2"}, nil)
		})
		It("rename sub level", func() {
			Expect(fs.Rename("/d1/d1n1", "d2")).To(Succeed())
			ExpectFolders(fs, "d2/", []string{"d1n1a"}, nil)
		})
		It("fail rename root", func() {
			Expect(fs.Rename("/", "d2")).To(Equal(errors.New("cannot rename root dir")))
		})
		It("fail rename to existent", func() {
			Expect(fs.Rename("/d1/d1n1", "d1/d1n2")).To(Equal(os.ErrExist))
		})
	})
})
