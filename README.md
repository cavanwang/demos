# demos
## All kinds of tiny program by golang.
### play-mp3
It plays a mp3 file on your computer using default audio device.
#### Build and run:
    go mod init && go mod tidy

    cd play-mp3
    go mod tidy
    CGO_ENABLED=1 go build
    // Following will play the test.mp3 under the directory.
    ./play-mp3

    cd lockdb
    go build
    // Following will create sqlite.db under the current directory and
    // show 2 goroutines preempting the lock.
    ./lockdb

    cd polish-notation
    go build
    // Following expression will got 21.0
    ./polish-notation '3+4/2 * (5+6/1.5)'
    
