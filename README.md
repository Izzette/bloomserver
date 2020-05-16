# bloomserver

Quick and dirty Golang bloom filter lookup server targeting leaked password
lookups.

# Quick-start guide

```shell
# Install packages
go get github.com/Izzette/bloomserver
go get github.com/Izzette/bloomserver/bloomserver-util

# Enter a new temorary directory
tmp="$(mktemp -d)"
cd "$tmp"

# Create a new bloom filter file (100000000 bits (~12MiB), 7 hash functions ≅ 1% false positive ratio)
~/go/bin/bloomserver-util -bloom-filter-file filter.bfdat new 100000000 7

# Grab your favorite leaked passwords list (one per line)
curl -LO https://github.com/danielmiessler/SecLists/raw/master/Passwords/Common-Credentials/10-million-password-list-top-1000000.txt

# Populate the bloom filter file with the passwords list
~/go/bin/bloomserver-util -bloom-filter-file filter.bfdat addAll 10-million-password-list-top-1000000.txt

# Start a server on http://127.0.0.1:14519
~/go/bin/bloomserver -bloom-filter-file filter.bfdat -listen-address tcp://127.0.0.1:14519 &
server_pid=$!

# Wait for the server to start
while ! curl -sf -d 'test' 'http://127.0.0.1:14519/api/search' > /dev/null; do
  sleep 1
done

# Test some naughty passwords
curl --data-raw 'password' 'http://127.0.0.1:14519/api/search' # => {"guiltySubstrings":["password"]}
curl --data-raw 'password' 'http://127.0.0.1:14519/api/search?substringLength=5' # => {"guiltySubstrings":["passw","passwo","passwor","password","asswor","assword","ssword","sword"]}

# Cleanup
kill "$server_pid"
cd -
rm -r "$tmp"
```

# License

Idk whatever, just do what you want.
