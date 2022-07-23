# !/bin/bash
rm result.txt

readSize=($(seq 1 100 4001))
 
for rs in ${readSize[@]}; do
    export ReadDataLength=$rs
    echo "Running with ReadDataLength=$ReadDataLength" >> result.txt
    echo "Running with ReadDataLength=$ReadDataLength" 
    go test -benchmem -run=^$ -bench=\^BenchmarkReadAndBuildPacket_Prependable\|BenchmarkReadAndBuildPacket_ByteSlice$ github.com/mucz/prependable -benchtime=1s >> result.txt
done

cat result.txt | grep "ReadAnd\|Running" > result_for_draw.txt

python3 draw.py