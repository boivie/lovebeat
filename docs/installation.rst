Installation
============

Using docker
------------

We keep up-to-date docker_ images. They are really tiny, too.

To get started, simply:

.. code-block:: bash

    $ docker run -it -p 8127:8127/udp -p 8127:8127/tcp -p 8080:8080 boivie/lovebeat

You may want to run with other options to specify volumes for the data and
configuration.

Prebuilt executable
-------------------

Our releases are built and uploaded to github. You can download a binary
matching your architecture and OS at:

  https://github.com/boivie/lovebeat/releases

Building from source
--------------------

You will need to have a go_ toolchain installed as well
as npm_ which will be used for downloading all other
dependencies for the frontend development.

After that, simply:


.. code-block:: bash

    $ mkdir go
    $ cd go
    $ export GOPATH=`pwd`
    $ go get github.com/boivie/lovebeat
    $ cd src/github.com/boivie/lovebeat
    $ make
    $ ./lovebeat

.. _go: http://golang.org
.. _npm: https://www.npmjs.com/
.. _docker: https://www.docker.com/
