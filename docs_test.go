package deployutility

import (
	"testing"

	"github.com/hafio/go-deploy-utility/notice"
	"github.com/hafio/go-deploy-utility/notice/noticetest"
)

// TestEveryDocCarriesTheSupportNotice: README.md and every doc under docs/ open
// with this module's support notice, directly under the title. The glob does not
// descend, so docs/handoff/ (planning history) is not held to it.
func TestEveryDocCarriesTheSupportNotice(t *testing.T) {
	n, err := notice.New("go-deploy-utility", "module", "the Solace products that the tools built on it deploy")
	if err != nil {
		t.Fatal(err)
	}
	noticetest.AssertDocs(t, n, ".", noticetest.Options{Strict: true})
}
