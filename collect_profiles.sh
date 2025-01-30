#!/bin/bash
# Collect memory profile
curl -o profiles/heap_server.pprof http://localhost:6060/debug/pprof/heap
# Collect CPU profile for 60 seconds
curl -o profiles/cpu_server.pprof http://localhost:6060/debug/pprof/profile\?seconds\=30