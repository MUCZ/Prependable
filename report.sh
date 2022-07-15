# !/bin/bash
rm result.txt
readSize=(1 10 20 30 40 50 60 70 80 90 100 
          200 300 400 500 600 700 800 900
          1000 1100 1200 1300 1400 1500 1600 1700 
          1800 1900 2000 2100 2200 2300 2400 2500 2600 2700
          2800 2900 3000 3100 3200 3300 3400 3500 3600 3700 
          3800 3900 4000) # make this auto 
 
for rs in ${readSize[@]}; do
    export RBL=$rs
    go test -benchmem -run=^$ -bench=.  github.com/mucz/prependable -benchtime=1s >> result.txt
done

cat result.txt | grep -v "goos" | grep -v "goarch" | grep -v "PASS" | grep -v "ok" |grep -v "cpu" | grep -v "pkg" > result.txt