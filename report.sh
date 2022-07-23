# !/bin/bash
rm result.txt
readSize=($(seq 1 500 4001))
 
for rs in ${readSize[@]}; do
    export RBL=$rs
    echo "Running with RBL=$RBL"
    go test -benchmem -run=^$ -bench=.  github.com/mucz/prependable -benchtime=1s >> result.txt
done

cat result.txt | grep -v "goos" | grep -v "goarch" | grep -v "PASS" | grep -v "ok" |grep -v "cpu" | grep -v "pkg" > result.txt
python3 draw.py