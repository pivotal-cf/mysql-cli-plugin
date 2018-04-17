package unpack_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-cli-plugin/unpack/unpackfakes"

	"github.com/gobuffalo/packr"
	. "github.com/pivotal-cf/mysql-cli-plugin/unpack"
)

var _ = Describe("unpacker", func() {
	var (
		fakeBox  packr.Box
		unpacker *Unpacker
	)

	BeforeEach(func() {
		fakeBox = packr.NewBox("./fixtures")
		unpacker = &Unpacker{Box: fakeBox}
	})

	It("Unpacks assets to a directory", func() {
		tmpDir, err := ioutil.TempDir(os.TempDir(), "test-tmp-dir_")

		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		err = unpacker.Unpack(tmpDir)

		Expect(err).NotTo(HaveOccurred())
		Expect(tmpDir).To(BeADirectory())
		Expect(filepath.Join(tmpDir, "fake-migrate")).To(BeARegularFile())
		Expect(filepath.Join(tmpDir, "fake-assets")).To(BeADirectory())
		Expect(filepath.Join(tmpDir, "fake-assets", "fake-mysqldump")).To(BeARegularFile())
	})

	Context("When an invalid destination directory is specified", func() {
		var badBox *unpackfakes.FakeBox
		BeforeEach(func() {
			badBox = &unpackfakes.FakeBox{}
			unpacker.Box = badBox
		})

		It("Returns an error", func() {
			badBox.WalkReturns(errors.New("some-error"))
			err := unpacker.Unpack("/")

			Expect(err).To(MatchError("Error extracting migrate assets: some-error"))
		})
	})
})
