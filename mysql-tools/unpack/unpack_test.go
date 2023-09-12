package unpack_test

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/unpack/unpackfakes"

	"github.com/gobuffalo/packr"
	. "github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/unpack"
)

var _ = Describe("Unpacker", func() {
	var (
		fakeBox  packr.Box
		unpacker *Unpacker
		tmpDir   string
	)

	BeforeEach(func() {
		fakeBox = packr.NewBox("./fixtures")
		unpacker = NewUnpacker()
		unpacker.Box = fakeBox

		var err error
		tmpDir, err = ioutil.TempDir(os.TempDir(), "test-tmp-dir_")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	It("Unpacks assets to a directory", func() {
		Expect(unpacker.Unpack(tmpDir)).To(Succeed())

		Expect(tmpDir).To(BeADirectory())
		Expect(filepath.Join(tmpDir, "fake-migrate")).To(BeARegularFile())
		Expect(filepath.Join(tmpDir, "fake-assets")).To(BeADirectory())
		Expect(filepath.Join(tmpDir, "fake-assets", "fake-mysqldump")).To(BeARegularFile())
	})

	Context("when unpacking fails", func() {
		var (
			fakeFilesystem *unpackfakes.FakeFilesystem
			fakeFile       *unpackfakes.FakeFile
		)

		BeforeEach(func() {
			fakeFile = new(unpackfakes.FakeFile)
			fakeFile.WriteStub = func(p []byte) (int, error) {
				return len(p), nil
			}
			fakeFile.ReadReturns(0, io.EOF)

			fakeFilesystem = new(unpackfakes.FakeFilesystem)
			fakeFilesystem.CreateReturns(fakeFile, nil)

			unpacker.Filesystem = fakeFilesystem
		})

		It("returns an error when creating a directory fails", func() {
			fakeFilesystem.MkdirAllReturns(errors.New("creating directory failed"))

			err := unpacker.Unpack(tmpDir)
			Expect(err).To(MatchError(`Error extracting migrate assets: creating directory failed`))
		})

		It("returns an error when creating a file fails", func() {
			fakeFilesystem.CreateReturns(nil, errors.New("open() failed"))

			err := unpacker.Unpack(tmpDir)
			Expect(err).To(MatchError(`Error extracting migrate assets: open() failed`))
		})

		It("returns an error when writing to a file fails", func() {
			fakeFile.WriteReturns(0, errors.New("serious io error"))

			err := unpacker.Unpack(tmpDir)
			Expect(err).To(MatchError(`Error extracting migrate assets: serious io error`))
		})

		It("returns an error when closing a file fails", func() {
			fakeFile.CloseReturns(errors.New("serious error on file close"))

			err := unpacker.Unpack(tmpDir)
			Expect(err).To(MatchError(`Error extracting migrate assets: serious error on file close`))
			// Allow an extra close in defer
			Expect(fakeFile.CloseCallCount()).To(Equal(2))
		})

		Context("when unpacking in a non-windows environment", func() {
			BeforeEach(func() {
				if runtime.GOOS == "windows" {
					Skip("chmod not run on a windows environment")
				}
			})
			It("returns an error if chmod fails", func() {
				fakeFile.ChmodReturns(errors.New("cannot chmod"))

				err := unpacker.Unpack(tmpDir)
				Expect(err).To(MatchError(`Error extracting migrate assets: cannot chmod`))
			})
		})
	})
})
