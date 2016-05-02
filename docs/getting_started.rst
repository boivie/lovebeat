Getting started
===============

Concept
-------

Services send "heartbeats" with regular intervals to lovebeat. If they for some
reason stop sending the heartbeats, lovebeat will react to this and update the
service state and possibly trigger alarms, such as sending e-mails.

So let's break it down a bit:

Services
--------

This is the name of the "thing" you want to monitor. You can choose anything
as the name, and it typically looks like "myapp.mailers.invoice" with periods
as the delimiter.

As you grow and have a lot of services to monitor, it's good to have some
sort of hierarchy.

States
------

A service can be in different states. **OK** is the state you want to keep
your services in. You can set a timeout so that the service will change state
into **ERROR** if the service hasn't issued a beat within that period of time.

Views
-----

A view shows a filtered subset of your services. You specify a matching pattern
and all services whose identifiers match this pattern will be part of
the view.

This is an example of a view called "backup-jobs" that match all servies
starting with "backup."

.. code-block:: ini

    [[views]]
    name = "backup-jobs"
    pattern = "backup.*"

The views will inherit state from their services. If all services are **OK**,
the view will be **OK**. But if any service is **ERROR** state, the view will
transition into the **ERROR** state.

Views can be automatically created based on the service names, which is a
powerful feature when your service names have a structure.

Say that you have an application running on three servers (alpha, beta and
delta), and the application provides two heartbeats, ".healthcheck" and
".background-job-1". The complete list of services will thus be:

 * application-name.alpha.healthcheck
 * application-name.alpha.background-job-1
 * application-name.beta.healthcheck
 * application-name.beta.background-job-1
 * application-name.delta.healthcheck
 * application-name.delta.background-job-1

By having a view configuration such as:

.. code-block:: ini

    [[views]]
    name = "server-$name"
    pattern = "application-name.$name.*"

You will then end up with three views, "server-alpha" including the services
"application-name.alpha.healthcheck" and "application-name.alpha.background-job-1"
and similar for "server-beta" and "server-delta".

Web UI
------

Just point your browser to http://localhost:8080/
