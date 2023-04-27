#!/bin/sh
timeStart=$(date "+%Y-%m-%d %H:%M:%S").$((`date "+%N"`/1000000))
echo "k8s $HOSTNAME start_pre_stop ${timeStart}" >>/proc/1/fd/1
kill -15 1
sleep 10
timeEnd=$(date "+%Y-%m-%d %H:%M:%S").$((`date "+%N"`/1000000))
echo "k8s $HOSTNAME end_pre_stop ${timeEnd}" >>/proc/1/fd/1