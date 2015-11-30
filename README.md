# protoc-gen-cleango
Generates more cleaner go code.

# Usage
```
# installation
go get -u github.com/jerrodrurik/protoc-gen-cleango

# simple usage
protoc --cleango_out=. your.proto

# include $GOPATH/src as import path.
protoc -I. -I$GOPATH/src --cleango_out=. your.proto
```

# License
MIT
