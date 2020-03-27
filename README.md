# Bas-monitor sdk

## if go build with meta info, like this, do not include space 
go build -v -ldflags "-X main.Meta=build_date=`date -u '+%Y-%m-%d_%H:%M:%S'`&build_revision=`git rev-parse HEAD`"
