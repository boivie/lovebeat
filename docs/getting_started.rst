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
your services in. You can set two timeouts, a *warning* and an *error*
timeout. The service will change state into **WARNING** or **ERROR** when the
timeouts have expired. Both timeouts are optional.

Views
-----

A view shows a filtered subset of your services. You specify a regular
expression and all services whose identifiers match this pattern will be part of
the view.

The views will inherit state from their services. If all services are **OK**,
the view will be **OK**. But if any service is in **WARNING** or **ERROR**
state, the view will transition into the **WARNING** or **ERROR** state.

Configuration
~~~~~~~~~~~~~

This is an example of a view called "backups" that match all servies starting
with "backup."

.. code-block:: ini

    [views.backups]
    pattern = "backup.*"

Web UI
------

Just point your browser to http://localhost:8080/
