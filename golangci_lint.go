 golangci-lint
on:
  push:
    paths:
      - "go.sum"
      - "go.mod"
      - "**.go"
      - "scripts/errcheck_excludes.txt"
      - ".github/workflows/golangci-lint.yml"
      - ".golangci.yml"
  pull_request:

jobs:
  golangci:
    name: lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v3
      - name: install Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.17.x
      - name: Lint
        uses: golangci/golangci-lint-action@v3.1.0
        with:
          version
staking-GMG/golangci-lint.yml at Main · GIMICI/staking-GMG
