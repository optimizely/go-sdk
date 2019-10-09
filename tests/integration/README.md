# Optimizely Go SDK Local FSC Testing

#### Setup
1. Check the `GOPATH` env variable.
2. Clone sdk repo under this path.
   `$GOPATH/src/github.com/optimizely/`
3. Open terminal and switch directory to the cloned sdk `$GOPATH/src/github.com/optimizely/go-sdk`.
4. Run the following command to fetch dependencies: <pre>``` go get ```</pre>
5. Set **DATAFILES_DIR='${Path to datafiles folder you want to use}'** Environment variable.
6. Copy all feature files to `$GOPATH/src/github.com/optimizely/go-sdk/tests/integration/features` folder.
7. Run the following command to execute gherkin tests: <pre>``` go test ./tests/integration/ ```</pre>

For further instructions: https://golang.org/doc/code.html
