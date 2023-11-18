#!/bin/bash

# Depends on godoc2markdown being installed.
# Installation instructions: https://git.sr.ht/~humaid/godoc2markdown

HOME_DIR=$PWD
dir=$1

cd $dir
title=$(echo "$dir" | tr '/' '_')
outfile=$HOME_DIR/docs/api/$title.md
echo "---" > $outfile
echo "title: $dir" >> $outfile
echo "---" >> $outfile
echo "# $dir" >> $outfile
go doc -all | godoc2markdown >> $outfile
cd $HOME_DIR
