module github.com/manics/go-conda-proxy

go 1.20

require (
	github.com/stretchr/testify v1.8.4
	golang.org/x/exp v0.0.0-20230801115018-d63ba01acd4b
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
)

replace github.com/manics/go-conda-proxy/repodata => ./repodata
