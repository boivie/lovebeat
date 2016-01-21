API
===

statsd API
----------

Using UDP or TCP.

We use the graphite_ protocol to trigger heartbeats and to set warning/error
timeouts. UDP is great for generating heartbeats from within your application
since the performance cost is very small and your application will not be affected
if the lovebeat server isn't running for any reason.

  * To trigger a heartbeat, send a counter value >= 0 to "<service>.beat"
  * To set a warning timeout (in seconds), set the gauge value of "<service>.warn"
  * To set an error timeout (in seconds), set the gauge value of "<service>.err"
  * To clear a value, set the timeout to -1.
  * To set the timeout to be automatically calculated, set the timeout to -2.
  * A shortcut for setting 'err' to -2 and issuing a beat (since this is a
    fairly common pattern), send a counter to "<service>.autobeat"

Examples:

.. code-block:: bash

    # UDP
    $ echo "invoice.mailer.beat:1|c" | nc -4u -w0 localhost 8127

    # TCP
    $ echo "invoice.mailer.warn:3600|g" | nc -c localhost 8127

    # TCP, setting warn to 'auto'
    $ echo -e "invoice.mailer.warn:-2|g\ninvoice.mailer.beat:1|c" | nc -c localhost 8127

    # UDP, sending a beat and auto-generating an error threshold
    $ echo "invoice.mailer.autobeat:1|c" | nc -4u -w0 localhost 8127

You can even put a statsd proxy in front of lovebeat if you don't want to send
UDP packets outside your localhost.

HTTP API
--------

The HTTP API is the easy way to send heartbeats from e.g. curl.

This API is also used by the web UI and is fairly complete.

POST /api/services/<service_id>
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Generates a heartbeat.

  * To set a warning timeout, add the form field "warn-tmo" and specify the
    time in seconds or "auto" to calculate one.
  * To set an error timeout, add the form field "err-tmo" and specify the
    time in seconds or "auto" to calculate one.

This endpoint returns an empty JSON object as response.

GET /api/services
~~~~~~~~~~~~~~~~~

Returns the list of services.

  * By specifying a `?view=<name>` query parameter, only services that are
    members of the specified view will be returned.

GET /api/services/<service_name>
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Returns information about a specific service.

* By setting the `?details=1` query parameter, additional information may
  be returned.

DELETE /api/services/<service_name>
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Deletes a service.

GET /api/views
~~~~~~~~~~~~~~

Returns a list of views.

GET /api/views/<view_name>
~~~~~~~~~~~~~~~~~~~~~~~~~~

Returns details of a specific view.

Examples:

.. code-block:: bash

    $ curl -X POST http://localhost:8080/api/services/invoice.mailer
    $ curl -d err-tmo=1800 http://localhost:8080/api/services/invoice.mailer

.. _graphite: http://graphite.wikidot.com/
