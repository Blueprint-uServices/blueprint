#!/bin/bash
HOME_DIR=$PWD

# Remove all the old docs
mkdir -p $HOME_DIR/docs/api
cd $HOME_DIR/docs/api/
rm *.md
cd $HOME_DIR

