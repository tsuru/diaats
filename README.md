#diaats

Docker-image-as-a-tsuru-service is a tool for generating a [tsuru service
API](http://docs.tsuru.io/en/stable/services/api.html) based on a Docker image.

Each registered Docker image is a plan, and they're defined in the IMAGE_PLANS
environment variable. The value of this variable must be a valid JSON. For
example:

```
IMAGE_PLANS='[{"image":"elasticsearch","plan":"elasticsearch"},{"image":"memcached","plan":"memcached"},{"image":"registry.mycompany.com/team/memcached:1.4,"plan":"custom_memcached"}]'
```

Other relevant environment variables include:

 - DOCKER_HOST: the address of the host. It might be a Swarm cluster. This
   setting is mandatory.
 - DOCKER_CONFIG: a JSON representing a
   [HostConfig](https://godoc.org/github.com/fsouza/go-dockerclient#HostConfig)
   instance. This environment variable is optional.
 - API_USERNAME and API_PASSWORD: in case the user wants to enable basic
   authentication in the API, these environment variables must be defined. They
   might be omitted, which means no authentication.
 - MONGODB_URL: the [MongoDB connection
   string](http://docs.mongodb.org/manual/reference/connection-string/). The
   API will use MongoDB to store metadata about the instances in the service.
   This setting is mandatory.

What the API does:

 - on service-add, it creates a container on the configured Docker host
 - on service-bind, it returns a list of endpoints in the format
   [host_ip]:[host_port], for each port exported by the Docker image
 - on service-unbind, it doesn't do anything
 - on service-remove, it removes the container from the configured Docker host

##Deployment example

Users could deploy this API as a "memcached" service, offering multiple
versions of memcached, as available in the [Docker
Hub](https://registry.hub.docker.com/_/memcached/tags/manage/). Each version
will be a plan, and each container will have 256 MB of RAM memory. The
configuration for doing so would be:

```
IMAGE_PLANS='[{"image":"memcached:1","plan":"memcached_1"},{"image":"memcached:1.3","plan":"memcached_1_4"},{"image":"memcached:1.4.21","plan":"memcached_1_4_21"},{"image":"memcached:1.4.22","plan":"memcached_1_4_22"},{"image":"memcached:1.4.23","plan":"memcached_1_4_23"},{"image":"memcached:1.4.24","plan":"memcached_1_4_24"}]'
DOCKER_HOST='tcp://10.10.10.10:2376'
DOCKER_CONFIG='{"Memory":268435456}'
```
