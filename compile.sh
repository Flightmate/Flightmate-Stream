# assign execute permissions with chmod +x compile.sh before running with ./compile.sh
env GOOS=windows GOARCH=386 go build -o executables/streamclient-windows-386.exe client.go &
env GOOS=windows GOARCH=amd64 go build -o executables/streamclient-windows.exe client.go &
env GOOS=darwin GOARCH=amd64 go build -o executables/streamclient-macOS client.go &
env GOOS=linux GOARCH=386 go build -o executables/streamclient-linux-386 client.go & 
env GOOS=linux GOARCH=amd64 go build -o executables/streamclient-linux client.go