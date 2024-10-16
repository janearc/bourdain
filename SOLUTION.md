# solution

A Makefile exists with a variety of targets. 

The first thing you'll want to do is `make initdb`, which generates test data,
generates the schema, installs the helper plsql functions, and loads the data.

Note that this is going to require docker-compose and some kind of docker
environment on your build machine.

After you have initialized the database and fed it some data, you can then
run the "solution" with `make runCheckAvailability`.

(to just do all the things: `make checks`)

# goodies

You can run `make names` to see the example names which are honestly hilarious.

Assorted things are available for poking around:
* `make dblogs`
* `make applogs`
* `make psql`
* `make stats`

# design overview

docker-compose is used to create a two-stage docker image that includes a postgres
instance and the tooling used to initialize and test the database. web services are
handled by golang. `main.go` lives in `service` and is complicated to build, so you'll
want to do that with the Makefile (`make buildWebService`). running the service is 
likewise complicated because it is inside docker, but binaries are generated in the 
root directory and you should be able to run them directly.

the problem is complicated because of the nature of transactional state. when one requests
a reservation, the first problem is "who is this reservation for," followed by "what are their needs,"
followed by "given the needs of the party, what restaurants are available, in the time requested?"

so the way we do this is by offloading most of this to the database. this provides a
unified mechanism for accessing the database, lets the database do the things the database
is good at, and it simplifies the golang code (aside from the part which actually
generates the schema and so forth, which *could* have been done in shell or similar, 
but I didn't really feel like having a complicated set of shell stuff on top of a 
makefile, a bunch of sql stuff, and go stuff. there's a lot of different environments here)

you can think of the environment as being separated into a few distinct parts:

- your environment (your laptop, whatever; this is mostly Makefile)
- docker-compose 
- docker images 
- the web service (golang)
- the database (postgres)
- the database schema (tables)
- the stored procedures (plsql functions that serve as a kind of api to the golang stuff)
- the data (which is both the "data" and the transactional state)

this seems pretty heavy-handed for two endpoints, and it is. however, because this is
in docker, it is pretty trivial to just dump this into kubernetes or whatever, and host
in the cloud. because much of the work is done with stored procedures, adding additional 
fields is not world-ending for golang, and performance can be addressed with schema
changes and procedure changes in the database. and lastly, because this application is
structured this way, extending it is pretty trivial. relatively speaking.

# retrospective thoughts

this could have been done in a much simpler way: shell stuff to initialize the
database, no real need for makefiles, could have been done in sqllite, could have 
been done with all the transactional logic and state-keeping in memory in golang.
it would have been smaller and easier to write.

but this kind of simple, just-enough-to-clear-the-bar sort of solution is not the
kind of solution that lets me sleep well at night. i want something that i can 
extend by just adding a new file and copying some stub code out of the repository 
i already have. this front-loads a lot of effort at starting the project, but it also
means that when my manager inevitably says "well, what if we wanted to base requests
for restaurant availability on the phase of the moon," it's not going to take me two weeks
while i refactor everything.

thus, design choices were made which may seem counter-intuitive, but intuition
depends on perspective and experience.

also the location/gis/mapping problem is interesting and fun but it turns out that
building an image on my arm64 laptop against a docker image that is linux amd64 is
a real hassle, and you wind up building a docker image entirely just building
postgres and gis stuff from scratch and the build takes like 20 minutes, and ultimately
it is just not worth it for this application. however, i know how that would be done,
and honestly i love building systems from scratch so that might just be something
i put in my back pocket for the future.