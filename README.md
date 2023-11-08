# demos
## All kinds of tiny program by golang.
### play-mp3
It plays a mp3 file on your computer using default audio device.
#### Build and run:
    cd play-mp3
    go mod init && go mod tidy
    CGO_ENABLED=1 go build

    // Following will play the test.mp3 under the directory.
    ./play-mp3
