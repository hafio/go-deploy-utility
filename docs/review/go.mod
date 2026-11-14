// Not Go code. This file makes docs/review/ a module of its own, so the Go module
// zip of github.com/hafio/go-deploy-utility leaves the review reports and their
// findings out: a consumer that fetches a tag does not download them. The
// directory stays in git.
module github.com/hafio/go-deploy-utility/docs/review

go 1.27
