
#!/bin/bash

set -e

files=$(find . -type f -name '*.go')
exitcode=0
for f in $files
do
    cmd="gofmt -e -l $f | wc -l"
    res=$(eval $cmd)
    if [ $res -gt 0 ]
    then
        echo "cmd: \"$cmd\" failed. cmd result = $res"
        exitcode=1
    fi
done

exit $exitcode