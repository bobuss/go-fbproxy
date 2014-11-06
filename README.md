# Facebook Service Proxy : proof of concept

Yep, I already have a similar service, working in production, based on a python/uwsgi stack.

This one consists on my first GO experiments...

- net/http
- mux router
- redis



## Howto

Run the service : it reponses on the port 3000.

    $ ./server


Test it with the little python (3.4+) test.py.

    $ python3.4 test.py