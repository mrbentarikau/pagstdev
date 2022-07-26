#!/bin/sh

RESTART='/home/caubert/go/src/github.com/jonas747/yagpdb/cmd/yagpdb/yagpdb -web -bot'
PGREP="/bin/pgrep"
YAGPDB="yagpdb"

$PGREP $YAGPDB

if [ $? -ne 0 ]
then
    tmux detach
    tmux a
    tmux select-window -t PAGSTDB
    tmux send 'doas /home/caubert/go/yagpdb/cmd/yagpdb/yagpdb -all -backgroundworkers -pa -syslog' ENTER
fi
