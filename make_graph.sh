#!/bin/bash
go tool pprof -raw -output=out.txt cpu_profile.prof
/home/john/FlameGraph/stackcollapse-go.pl ./out.txt | /home/john/FlameGraph/flamegraph.pl > flame.svg
mv ./flame.svg /mnt/c/Users/JohnD/Desktop/flame.svg

