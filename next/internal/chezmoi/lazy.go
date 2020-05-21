package chezmoi

import (
	"crypto/sha256"
)

type contentsFunc func() ([]byte, error)

// A lazyContents evaluates its contents lazily.
type lazyContents struct {
	contentsFunc   contentsFunc
	contents       []byte
	contentsErr    error
	contentsSHA256 []byte
}

type linknameFunc func() (string, error)

// A lazyLinkname evaluates its linkname lazily.
type lazyLinkname struct {
	linknameFunc linknameFunc
	linkname     string
	linknameErr  error
}

// newLazyContents returns a new lazyContents with contents.
func newLazyContents(contents []byte) *lazyContents {
	return &lazyContents{
		contents: contents,
	}
}

// Contents returns e's contents.
func (lc *lazyContents) Contents() ([]byte, error) {
	if lc == nil {
		return nil, nil
	}
	if lc.contentsFunc != nil {
		lc.contents, lc.contentsErr = lc.contentsFunc()
		lc.contentsFunc = nil
		if lc.contentsErr == nil {
			lc.contentsSHA256 = sha256Sum(lc.contents)
		}
	}
	return lc.contents, lc.contentsErr
}

// ContentsSHA256 returns the SHA256 sum of f's contents.
func (lc *lazyContents) ContentsSHA256() ([]byte, error) {
	if lc == nil {
		return sha256Sum(nil), nil
	}
	if lc.contentsSHA256 == nil {
		contents, err := lc.Contents()
		if err != nil {
			return nil, err
		}
		lc.contentsSHA256 = sha256Sum(contents)
	}
	return lc.contentsSHA256, nil
}

// newLazyLinkname returns a new lazyLinkname with linkname.
func newLazyLinkname(linkname string) *lazyLinkname {
	return &lazyLinkname{
		linkname: linkname,
	}
}

// Linkname returns s's linkname.
func (ll *lazyLinkname) Linkname() (string, error) {
	if ll == nil {
		return "", nil
	}
	if ll.linknameFunc != nil {
		ll.linkname, ll.linknameErr = ll.linknameFunc()
		ll.linknameFunc = nil
	}
	return ll.linkname, ll.linknameErr
}

func sha256Sum(data []byte) []byte {
	sha256SumArr := sha256.Sum256(data)
	return sha256SumArr[:]
}
