// Not Go code. This file makes graphify-out/ a module of its own, so the Go
// module zip of github.com/hafio/go-deploy-utility leaves the knowledge graph out:
// a consumer that fetches a tag does not download it. The directory stays in git.
module github.com/hafio/go-deploy-utility/graphify-out

go 1.27
