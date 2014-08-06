This is a work in progress.

Example:

echo "foo.beat:60|c" | nc -u -w 0 localhost 8127

Beats should be counters, and warn/err should be gauges

It's best to use gauges instead of counters as the values will be left untouched
in case you have a statsd proxy between the client and the lovebeat server.

