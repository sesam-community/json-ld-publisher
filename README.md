
# Json-LD Publisher Service
Sesam Microservice that exposes datasets as JSON-LD

# Build

This service is written in Go.

<pre>
git clone git@github.com:sesam-community/json-ld-publisher.git
cd src
go build
</pre>

To Build the docker image:

<pre>
cd src
docker build -t tag-name .
</pre>

## Configure the Service

The service requires the following environment variables to be set:

<pre>
SESAM_JWT="Service Account JWT for accessing the datasets to be exposed."

SESAM_API="The URL of the Sesam Instance API"

SERVICE_PORT="Port on which service runs in container. Default is 5000. "

SESAM_DATASETS="The names the datasets to exposse ';' seperated. e.g: people;products"

"GIN_MODE" : "release or debug. Use release in production"

"JSON_LD_SERVICE_API" : "The url of the public facing endpoint of this microservice",

"JSONLD_MODE" : "CONTEXT_REF_INLINE or CONTEXT_FIRST. Defines the mode of how the JSON-LD context is delivered."
</pre>

This service supports the ```since``` parameter as defined in the Sesam JSON Pull protocol.

## Calling the service

There are two endpoints exposed:

<pre>

/context => returns the JSON-LD context representation for that Sesam instance. This is derived from having Namespaces defined in the Sesam Service Config.

/datasets/{dataset-id} => returns a streamed array of Sesam JSON-LD entities. If CONTEXT_REF_INLINE is used each object has a @context property that links to the context URL. If CONTEXT_FIRST is used then the first object in the array contains just the context. 

</pre>

## Logging

The service logs to stdout.
