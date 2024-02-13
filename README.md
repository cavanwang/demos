# demos
## All kinds of tiny program by golang.
### play-mp3
It plays a mp3 file on your computer using default audio device.
### lockdb
It tries to lock a name with timeout, like mysql's GET_LOCK(name, timeout)
### polish-notation
It evals an arithmetic express by reverse polish notation
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
    // Following expression will get 21.0
    ./polish-notation '3+4/2 * (5+6/1.5)'

    cd compilation-principle/arithmetic-expression/left-recursive-correct-associativity
    go build
    // Following will use special left-recursive to resolve arithmetic express
    ./left-recursive-correct-associativity

    cd compilation-principle/arithmetic-expression/left-recursive-error-associativity
    go build
    // Following will use standard left-recursive to resolve expression and it cannot 
    // left-associativity, e.g. 3+4-5 will wrongly got a 3+(4-5)
    ./left-recursive-error-associativity
    
