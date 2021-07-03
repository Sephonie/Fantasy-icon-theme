/*
Package restful , a lean package for creating REST-style WebServices without magic.

WebServices and Routes

A WebService has a collection of Route objects that dispatch incoming Http Requests to a function calls.
Typically, a WebService has a root path (e.g. /users) and defines common MIME types for its routes.
WebServices must be added to a container (see below) in order to handler Http requests from a server.

A Route is defined by a HTTP method, an URL path and (optionally) the MIME types it consumes (Content-Type) and produces (Accept).
This package has the logic to find the best matching Route and if found, call its Function.

	ws := new(restful.WebService)
	ws.
		Path("/users").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/{user-id}").To(u.findUser))  // u is a UserResource

	...

	// GET http://localhost:8080/users/1
	func (u UserResource) findUser(request *restful.Request, response *restful.Response) {
		id := request.PathParameter("user-id")
		...
	}

The (*Request, *Response) arguments provide functions for reading information from the request and writing information back to the response.

See the example https://github.com/emicklei/go-restful/blob/master/examples/restful-user-resource.go with a full implementation.

Regular expression matching Routes

A Route parameter can be specified using the format "uri/{var[:regexp]}" or the special version "uri/{var:*}" for matching the tail of the path.
For example, /persons/{name:[A-Z][A-Z]} can be used to restrict values for the para