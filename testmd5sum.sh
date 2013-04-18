#!/bin/bash

TEST_FILE="testmd5sum.tmp"

if [ -n "$1" ]
then
  mnt=$1
  size=$2
else
  echo "Usage: testmd5sum.sh [mount location] [file size]"
  exit 0
fi

dd if=/dev/urandom of=$TEST_FILE bs=1M count=$size

a=$(cat $TEST_FILE | md5sum)
echo $a

mv $TEST_FILE $mnt
cd $mnt

b=$(cat $TEST_FILE | md5sum)
echo $b

if [ "$a" != "$b" ]
then
  echo "md5sum failed to verify integrity"
else
  echo "Successfully verified file integrity"
fi

rm -f $TEST_FILE
cd

exit 0
