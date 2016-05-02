API
===

statsd API
----------

Using UDP or TCP.

We use the graphite_ protocol to trigger heartbeats and to set timeouts.
UDP is great for generating heartbeats from within your application
since the performance cost is very small and your application will not be affected
if the lovebeat server isn't running for any reason.

  * To trigger a heartbeat, send a counter value >= 0 to "<service>.beat"
  * To set a timeout (in seconds), set the gauge value of "<service>.timeout"
  * To clear a value, set the timeout to -1.
  * To set the timeout to be automatically calculated, set the timeout to -2.
  * A shortcut for setting the timeout to -2 and issuing a beat (since this is a
    fairly common pattern), send a counter to "<service>.autobeat"

Examples:

.. code-block:: bash

    # UDP
    $ echo "invoice.mailer.beat:1|c" | nc -4u -w0 localhost 8127

    # TCP
    $ echo "invoice.mailer.timeout:3600|g" | nc -c localhost 8127

    # TCP, setting timeout to 'auto'
    $ echo -e "invoice.mailer.timeout:-2|g\ninvoice.mailer.beat:1|c" | nc -c localhost 8127

    # UDP, same as above
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

  * To set a timeout, add the form field "timeout" and specify the
    time in seconds or "auto" to calculate one.
  * You can also set the timeout value using a query parameter, e.g.
    ?timeout=3600
  * Last, but not least, you can also post a JSON payload to this endpoint
    and let the JSON object's timeout field be set to the timeout value. Note
    that the ``Content-Type`` must be set to ``application/json``.

This endpoint returns an empty JSON object as response.

Examples:

.. code-block:: bash

    # Only trigger a beat - don't set any value
    $ curl -X POST http://localhost:8080/api/services/invoice.mailer

    # Set the timeout using a form field value
    $ curl -d timeout=3600 http://localhost:8080/api/services/invoice.mailer

    # Set the timeout using a query parameter
    $ curl -X POST http://localhost:8080/api/services/invoice.mailer?timeout=3600

    # Setting the timeout as a JSON object.
    $ curl -H "Content-Type: application/json" -d '{"timeout":3600}' http://localhost:8080/api/services/invoice.mailer

GET /api/services
~~~~~~~~~~~~~~~~~

Returns the list of services.

  * By specifying a ``?view=<name>`` query parameter, only services that are
    members of the specified view will be returned.

GET /api/services/<service_name>
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Returns information about a specific service.

* By setting the ``?details=1`` query parameter, additional information may
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

.. _graphite: http://graphite.wikidot.com/
